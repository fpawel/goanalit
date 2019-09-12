package products

import (
	"github.com/boltdb/bolt"
	"github.com/fpawel/gutils/utils"
	"sort"
	"time"

	"log"
)

type times []time.Time

type logVisitor func(*bolt.Bucket, []byte, *LogRecord)

func (x times) Len() int {
	return len(x)
}

func (x times) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func (x times) Less(i, j int) bool {
	return x[i].Before(x[j])
}

func (x Logs) Times() (times times) {
	for k := range x {
		times = append(times, k)
	}
	sort.Sort(times)
	return
}

func (x Logs) Last() (t *time.Time, r *LogRecord) {
	times := x.Times()
	if len(times) > 0 {
		t = &times[len(times)-1]
		r = x[*t]
	}
	return
}

func (x LogRecord) ProductTime() (productTime ProductTime, ok bool) {
	for i := 2; i < len(x.Path)-1; i++ {
		if string(x.Path[i]) == "tests" && string(x.Path[i-2]) == "products" {
			productTime = ProductTime(KeyToTime(x.Path[i-1]))
			ok = true
			return
		}
	}
	return
}

func iterateBucketLogs(buck *bolt.Bucket, path [][]byte, visitor logVisitor) {

	c := buck.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		if len(k) < 8 {
			log.Panicf("bad log record key %s % X, %s % X", string(k), k, string(v), v)
			continue
		}
		if len(v) < 1 {
			log.Panicf("bad log record value %s % X, %s % X", string(k), k, string(v), v)
			continue
		}
		// time.Unix(0, u.BytesToInt64(k))
		visitor(buck, k, &LogRecord{
			path, int(v[0]), string(v[1:]),
		})
	}
}

func walkLogs(buck *bolt.Bucket, path [][]byte, f logVisitor) {
	if buck == nil {
		return
	}
	c := buck.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		if v == nil {
			if string(k) == "logs" {
				iterateBucketLogs(buck.Bucket(k), append(path, k), f)
			} else {
				walkLogs(buck.Bucket(k), append(path, k), f)
			}
		}
	}
}

func collectBucketLogs(buck *bolt.Bucket, logs Logs) Logs {

	if logs == nil {
		logs = make(Logs)
	}

	walkLogs(buck, nil, func(_ *bolt.Bucket, k []byte, r *LogRecord) {
		logs[time.Unix(0, utils.BytesToInt64(k))] = r
	})
	return logs
}

func bucketLogs2(buck *bolt.Bucket) (t1 time.Time, r1 *LogRecord, t2 time.Time, r2 *LogRecord) {

	var f bool
	walkLogs(buck, nil, func(_ *bolt.Bucket, k []byte, r *LogRecord) {
		t := time.Unix(0, utils.BytesToInt64(k))
		if f {
			if t.Before(t1) {
				t1 = t
				r1 = r
			} else if t.After(t2) {
				t2 = t
				r2 = r
			}
		} else {
			f = true
			t1 = t
			r1 = r
			t2 = t
			r2 = r
		}
	})
	return
}

func (x Tx) Logs(p [][]byte, logs Logs) Logs {
	buck := x.BucketRead(p)
	if buck == nil {
		return logs
	}
	return collectBucketLogs(buck, logs)
}

func (x Tx) Logs2(p [][]byte) (t1 time.Time, r1 *LogRecord, t2 time.Time, r2 *LogRecord) {
	buck := x.BucketRead(p)
	if buck == nil {
		return
	}
	t1, r1, t2, r2 = bucketLogs2(buck)
	return
}

func (x Tx) MostImportantLogRecord(p [][]byte) (l *LogRecord) {
	logs := x.Logs(p, nil)
	var t time.Time
	for rt, r := range logs {
		if l == nil || l.Level < r.Level || l.Level == r.Level && rt.After(t) {
			l = r
			t = rt
		}
	}
	return
}

func (x Tx) ClearLogs(p [][]byte) {
	buck := x.BucketWrite(p)
	k := []byte("logs")
	if buck.Bucket(k) != nil {
		if err := buck.DeleteBucket(k); err != nil {
			panic(err)
		}
		if err := buck.Put(k, nil); err != nil {
			panic(err)
		}
		if err := buck.Delete(k); err != nil {
			panic(err)
		}
	}
}

func (x Tx) DeleteLog(p [][]byte, key []byte) {
	buck := x.BucketWrite(append(p, []byte("logs")))
	if err := buck.Delete(key); err != nil {
		panic(err)
	}
}

func (x Tx) WriteLog(p [][]byte, timeKey []byte, level int, text string) []byte {
	p = append(p, []byte("logs"))
	buck := x.BucketWrite(p)
	value := append([]byte{byte(level)}, []byte(text)...)
	if timeKey == nil {
		timeKey = TimeToKey(time.Now())
	}
	if err := buck.Put(timeKey, value); err != nil {
		panic(err)
	}

	return timeKey
}

func (x Tx) WriteTestLog(p DBPath, r TestLogRecord) []byte {
	return x.WriteLog(Path(), r.TimeKey, r.Level, r.Text)
}

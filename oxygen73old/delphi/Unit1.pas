unit Unit1;

interface

uses
    Winapi.Windows, Winapi.Messages, System.SysUtils, System.Variants,
    System.Classes, Vcl.Graphics,
    Vcl.Controls, Vcl.Forms, Vcl.Dialogs, Vcl.StdCtrls, VclTee.TeeGDIPlus,
    VclTee.TeEngine, VclTee.Series, Vcl.ExtCtrls, VclTee.TeeProcs, VclTee.Chart,
    dateutils, System.SyncObjs,
    Vcl.Menus, Vcl.CheckLst, Vcl.ComCtrls, System.IniFiles,
    System.Generics.Collections;

type
    TForm1 = class(TForm)
        Panel1: TPanel;
        Panel2: TPanel;
        Panel3: TPanel;
        Panel4: TPanel;
        Panel5: TPanel;
        Chart1: TChart;
        Panel6: TPanel;
        RadioGroup1: TRadioGroup;
        Panel7: TPanel;
        ListView1: TListView;
        MainMenu1: TMainMenu;
        N1: TMenuItem;
        N2: TMenuItem;
        N3: TMenuItem;
        procedure FormCreate(Sender: TObject);
        procedure ListView1CustomDrawSubItem(Sender: TCustomListView;
          Item: TListItem; SubItem: Integer; State: TCustomDrawState;
          var DefaultDraw: Boolean);
        procedure ListView1SelectItem(Sender: TObject; Item: TListItem;
          Selected: Boolean);
        procedure FormActivate(Sender: TObject);
        procedure RadioGroup1Click(Sender: TObject);
        procedure N3Click(Sender: TObject);
        procedure N2Click(Sender: TObject);
        procedure ListView1ItemChecked(Sender: TObject; Item: TListItem);
    private
        { Private declarations }

    public
        { Public declarations }
        axisPress, axisPress1: TChartAxis;
        series_to_listitem: TDictionary<TChartSeries, TListItem>;
        listitem_to_series: TDictionary<TListItem, TChartSeries>;

        procedure SafeClose();
        procedure addXY(series_index: Integer; value_x: tdatetime;
          value_y: double);
        procedure updateSeriesListItem(series_index: Integer; value_y: double);
        procedure SetupListItem(listItem: TListItem; ser: TChartSeries);
        procedure addListItem(ser: TChartSeries);
        procedure insertListItem(index: Integer; ser: TChartSeries);
        procedure refresh_series_checked();
        procedure ClearSeries(N1: Integer; N2: Integer);

    end;

var
    Form1: TForm1;

type

    TReadPipeThread = class(TThread)

    private
        procedure Execute; override;
        procedure Handle_SeriesPoint(data: array of byte; restore:boolean);

    end;

implementation

{$R *.dfm}

uses utils;

const
    LINE_LEN = 80;
    MAX_WRITE = 1000 * sizeof(WideChar);
    IN_BUF_SIZE = 1000;

    CMD_ADD_CURRENT_TIME_VALUE = 1;
    CMD_SET_RADIOGROUP1_ITEMINDEX = 2;
    CMD_RESTORE_TIME_VALUE = 3;

var
    readPipeThread: TReadPipeThread;

    unix_begin_time: tdatetime;
    exe_dir: string;
    hPipe: THANDLE;
    lockPipe: TCriticalSection;
    iniFile: TIniFile;

procedure terminate_error(error_text: string);
var
    f: TextFile;
begin
    AssignFile(f, exe_dir + '\oxychart-fail.txt');
    ReWrite(f);
    WriteLn(f, error_text);
    CloseFile(f);
    ExitProcess(1);

end;

procedure writePipe(data: array of byte);
var
    writen_count: DWORD;
begin
    lockPipe.Acquire;
    if not(WriteFile(hPipe, data, length(data), writen_count, nil)) then
        terminate_error('error writing pipe');

    lockPipe.Release;
end;

procedure TForm1.FormActivate(Sender: TObject);
var
    wp: WINDOWPLACEMENT;
    window_placement_File: File of WINDOWPLACEMENT;
begin
    if FileExists(exe_dir + '\oxychart_position.') then
    begin
        AssignFile(window_placement_File, exe_dir + '\oxychart_position.');
        FileMode := fmOpenRead;
        Reset(window_placement_File);
        Read(window_placement_File, wp);

        SetWindowPlacement(Handle, wp);
        CloseFile(window_placement_File);

    end;
    self.OnActivate := nil;
end;

procedure TForm1.SafeClose();
var
    wp: WINDOWPLACEMENT;
    window_placement_File: File of WINDOWPLACEMENT;

begin
    if not GetWindowPlacement(Handle, wp) then
        terminate_error('GetWindowPlacement: false');

    AssignFile(window_placement_File, exe_dir + '\oxychart_position.');
    ReWrite(window_placement_File);
    Write(window_placement_File, wp);
    CloseFile(window_placement_File);
    self.close();
    // Application.Terminate;

end;

procedure TForm1.FormCreate(Sender: TObject);
var
    ser: TFastLineSeries;
    i: Integer;

begin
    // DeleteMenu(GetSystemMenu(Handle, false), SC_CLOSE, MF_BYCOMMAND);

    unix_begin_time := EncodeDateTime(1970, 1, 1, 0, 0, 0, 0);
    exe_dir := ExtractFileDir(paramstr(0));
    lockPipe := TCriticalSection.Create;
    iniFile := TIniFile.Create(exe_dir + '\oxychart.ini');

    series_to_listitem := TDictionary<TChartSeries, TListItem>.Create;
    listitem_to_series := TDictionary<TListItem, TChartSeries>.Create;

    axisPress := TChartAxis.Create(Chart1);
    axisPress.OtherSide := true;
    axisPress.PositionUnits := muPixels;
    axisPress.PositionPercent := -50;

    axisPress1 := TChartAxis.Create(Chart1);
    axisPress1.OtherSide := true;
    axisPress1.PositionUnits := muPixels;
    axisPress1.PositionPercent := -50;

    ser := TFastLineSeries.Create(nil);
    ser.XValues.DateTime := true;
    ser.Title := 'T,"C';
    ser.VertAxis := aRightAxis;
    Chart1.AddSeries(ser);

    ser := TFastLineSeries.Create(nil);
    ser.XValues.DateTime := true;
    ser.Title := 'P,μμ';
    ser.VertAxis := aRightAxis;
    ser.CustomVertAxis := axisPress;
    Chart1.AddSeries(ser);

    for i := 1 to 50 do
    begin
        ser := TFastLineSeries.Create(nil);
        ser.XValues.DateTime := true;
        ser.Title := Format('%02d', [i]);
        Chart1.AddSeries(ser);

    end;

    ser := TFastLineSeries.Create(nil);
    ser.active := false;
    ser.XValues.DateTime := true;
    ser.Title := 'T,"C';
    ser.VertAxis := aRightAxis;
    Chart1.AddSeries(ser);

    ser := TFastLineSeries.Create(nil);
    ser.active := false;
    ser.XValues.DateTime := true;
    ser.Title := 'P,μμ';
    ser.VertAxis := aRightAxis;
    ser.CustomVertAxis := axisPress1;
    Chart1.AddSeries(ser);

    for i := 1 to 50 do
    begin
        ser := TFastLineSeries.Create(nil);
        ser.XValues.DateTime := true;
        ser.Title := Format('%02d', [i]);
        ser.active := false;
        Chart1.AddSeries(ser);
    end;

    hPipe := CreateFileW(PWideChar('\\.\pipe\$Oxygen73Chart$'), GENERIC_READ or
      GENERIC_WRITE, FILE_SHARE_READ or FILE_SHARE_WRITE, nil,
      OPEN_EXISTING, 0, 0);
    if hPipe = INVALID_HANDLE_VALUE then
        terminate_error('hPipe = INVALID_HANDLE_VALUE');

    readPipeThread := TReadPipeThread.Create();

end;

procedure TForm1.ListView1CustomDrawSubItem(Sender: TCustomListView;
  Item: TListItem; SubItem: Integer; State: TCustomDrawState;
  var DefaultDraw: Boolean);
var
    r: Trect;
    c: tcanvas;
    i, d: Integer;
    ser: TChartSeries;

begin
    DefaultDraw := false;
    c := ListView1.Canvas;
    r := Item.DisplayRect(drBounds);
    for i := 0 to SubItem - 1 do
    begin
        r.Left := r.Left + ListView1.Columns.items[i].Width;
        r.Right := r.Left + ListView1.Columns.items[i + 1].Width;
    end;
    if Item.index < 0 then
        exit;

    assert(listitem_to_series.ContainsKey(Item));
    ser := listitem_to_series[Item];

    if SubItem = 1 then
    begin
        d := round(r.Top + r.Height / 2);
        c.Brush.Color := ser.SeriesColor;
        c.FillRect(Rect(r.Left, d - 2, r.Right, d + 2));
    end;

    if (SubItem = 2) and (ser.YValues.Count > 0) then
    begin
        c.Font.Size := 8;
        c.Refresh;
        DefaultDraw := true;
    end
    else
    begin
        if SubItem < 1 then
        begin
            Sender.Canvas.MoveTo(r.Left, r.Top);
            Sender.Canvas.LineTo(r.Right - 1, r.Bottom - 1);
        end;
        SetBkMode(Sender.Canvas.Handle, TRANSPARENT);

    end;
end;

procedure TForm1.ListView1ItemChecked(Sender: TObject; Item: TListItem);
begin
    if listitem_to_series.ContainsKey(Item) then
    begin
        listitem_to_series[Item].active := Item.Checked;
        refresh_series_checked();
    end;

end;

procedure TForm1.ListView1SelectItem(Sender: TObject; Item: TListItem;
  Selected: Boolean);
var
    i: Integer;
    ser: TFastLineSeries;
begin
    for i := 0 to Chart1.SeriesCount - 1 do
    begin
        ser := Chart1.Series[i] as TFastLineSeries;
        if ser = listitem_to_series[Item] then
        begin
            ser.LinePen.Width := 3;
        end
        else
        begin
            ser.LinePen.Width := 1;

        end;
    end;
end;

procedure TForm1.N2Click(Sender: TObject);
begin
    ClearSeries(52, 51 + 52);
end;

procedure TForm1.N3Click(Sender: TObject);
begin
    ClearSeries(0, 51);
end;

procedure TForm1.RadioGroup1Click(Sender: TObject);
var
    N1, N2: Integer;
    i: Integer;
    listItem: TListItem;
    ser: TChartSeries;
begin
    N1 := 0;
    N2 := 51;
    if RadioGroup1.ItemIndex = 1 then
    begin
        N1 := N1 + 52;
        N2 := N2 + 52;
    end;
    ListView1.Visible := false;
    ListView1.items.Clear;
    series_to_listitem.Clear;
    listitem_to_series.Clear;

    for i := 0 to Chart1.SeriesCount - 1 do
    begin
        Chart1.Series[i].active := false;
    end;

    for i := N1 to N2 do
    begin
        ser := Chart1.Series[i];
        ser.active := ser.XValues.Count > 0;
        if not ser.active then
            Continue;

        listItem := Form1.ListView1.items.Add;
        listItem.Caption := ser.Title;
        listItem.Checked := true;
        listItem.SubItems.Add('');
        listItem.SubItems.Add('');
        listitem_to_series.Add(listItem, ser);
        series_to_listitem.Add(ser, listItem);
    end;

    ListView1.Visible := true;
    refresh_series_checked;

end;

procedure TForm1.refresh_series_checked();
var
    active_t, active_p1, active_p2: Boolean;

begin

    axisPress.Visible := Chart1.Series[1].active;
    axisPress1.Visible := Chart1.Series[53].active;

    active_t := Chart1.Series[0].active or Chart1.Series[52].active;
    active_p1 := Chart1.Series[1].active;
    active_p2 := Chart1.Series[53].active;

    if not active_p1 and not active_p2 and not active_t then
    begin
        Chart1.MarginRight := 5;
        exit;
    end;

    Chart1.MarginRight := 50;

    if active_p1 and active_t then
    begin
        axisPress.PositionPercent := -50;
        exit;
    end;

    if active_p2 and active_t then
    begin
        axisPress1.PositionPercent := -50;
        exit;
    end;

    axisPress.PositionPercent := 0;
    axisPress1.PositionPercent := 0;
end;

procedure TForm1.addListItem(ser: TChartSeries);
var
    listItem: TListItem;
begin
    ListView1.Visible := false;
    listItem := ListView1.items.Add;
    SetupListItem(listItem, ser);
    ListView1.Visible := true;
end;

procedure TForm1.insertListItem(index: Integer; ser: TChartSeries);
var
    listItem: TListItem;
begin
    ListView1.Visible := false;
    listItem := ListView1.items.Insert(index);
    SetupListItem(listItem, ser);
    ListView1.Visible := true;
end;

procedure TForm1.SetupListItem(listItem: TListItem; ser: TChartSeries);
begin
    ListView1.Visible := false;
    listItem.Caption := ser.Title;
    listItem.Checked := ser.active;
    listItem.SubItems.Add('');
    listItem.SubItems.Add('');
    listitem_to_series.Add(listItem, ser);
    series_to_listitem.Add(ser, listItem);
    ListView1.Visible := true;
end;

procedure TForm1.ClearSeries(N1: Integer; N2: Integer);
var
    i: Integer;
begin
    for i := N1 to N2 do
    begin
        Chart1.SeriesList[i].Clear;
    end;

end;

procedure TForm1.updateSeriesListItem(series_index: Integer; value_y: double);
var
    listitem_inserted: Boolean;
    ser: TChartSeries;
    i: Integer;
    listItem: TListItem;
begin
    ser := Chart1.Series[series_index];
    if not series_to_listitem.ContainsKey(ser) then
    begin
        if ListView1.items.Count = 0 then
        begin
            addListItem(ser);

        end
        else
        begin
            listitem_inserted := false;
            for i := 0 to ListView1.items.Count - 1 do
            begin
                if series_index < listitem_to_series[ListView1.items[i]].SeriesIndex
                then
                begin
                    insertListItem(i, ser);
                    listitem_inserted := true;
                    break;
                end;
            end;
            if not listitem_inserted then
                addListItem(ser);
        end;
    end;

    listItem := series_to_listitem[ser];
    listItem.SubItems[1] := FormatFloat('.###', value_y);
    listItem.Update;

end;

procedure TForm1.addXY(series_index: Integer; value_x: tdatetime;
  value_y: double);
var
    i: Integer;
    x_millis: int64;
    listItem: TListItem;
    ser: TChartSeries;
    listitem_inserted: Boolean;
begin
    if series_index >= Chart1.SeriesCount then
    begin
        terminate_error('series index out of range');
    end;

    ser := Chart1.Series[series_index];

    ser.addXY(value_x, value_y);

    if (series_index < 52) and (RadioGroup1.ItemIndex = 1) or
      (series_index > 51) and (RadioGroup1.ItemIndex = 0) then
        exit;

    if ser.YValues.Count = 1 then
        ser.active := true;

end;

procedure TReadPipeThread.Execute;
var
    cmd_code: byte;

    readed_count: DWORD;
    buf: array [0 .. 2000] of byte;
    data: array of byte;
    data_bytes_count: longword;
    i: Integer;

begin
    while not Terminated do
    begin
        lockPipe.Acquire;
        if (not ReadFile(hPipe, buf, 5, readed_count, nil)) or
          (readed_count <> 5) then
        begin
            Synchronize(Form1.SafeClose);
            exit;
        end;
        lockPipe.Release;

        data_bytes_count := PUINT(@buf[1])^;
        cmd_code := buf[0];

        lockPipe.Acquire;
        if (not ReadFile(hPipe, buf, data_bytes_count, readed_count, nil)) or
          (readed_count <> data_bytes_count) then
        begin
            Synchronize(Form1.SafeClose);
            exit;
        end;
        lockPipe.Release;

        SetLength(data, data_bytes_count);
        for i := 0 to data_bytes_count - 1 do
            data[i] := buf[i];

        case cmd_code of
            CMD_ADD_CURRENT_TIME_VALUE:
                Handle_SeriesPoint(data, false);
            CMD_RESTORE_TIME_VALUE:
                Handle_SeriesPoint(data, true);
            CMD_SET_RADIOGROUP1_ITEMINDEX:
                Synchronize(
                    procedure
                    begin
                        Form1.RadioGroup1.ItemIndex := data[0];
                    end);

        else
            terminate_error(Format('unknown comand code %d', [cmd_code]));
        end;
        Finalize(data);

    end;

end;

procedure TReadPipeThread.Handle_SeriesPoint(data: array of byte; restore:boolean);
var
    series_index: byte;
    x_millis: int64;
    value_x: tdatetime;
    value_y: Single;

begin
    series_index := data[0];
    x_millis := PLONG64(@data[1])^;
    value_x := IncHour(IncMilliSecond(unix_begin_time, x_millis), 3);
    value_y := Psingle(@data[9])^;

    Synchronize(
        procedure
        begin
            Form1.addXY(series_index, value_x, value_y);
            if restore then
                Form1.RadioGroup1.ItemIndex := 0

            else
                Form1.updateSeriesListItem(series_index, value_y);


        end);
end;

end.

unit utils;

interface

uses
  System.SysUtils;

type
  TEnumHelper<T> = record
    class function Parse(x: String): T; static;
    class function Format(x: T): String; static;
  end;

implementation

uses
   System.TypInfo;


class function TEnumHelper<T>.Parse(x: String): T;
begin
  case Sizeof(T) of
    1: PByte(@Result)^ := GetEnumValue(TypeInfo(T), x);
    2: PWord(@Result)^ := GetEnumValue(TypeInfo(T), x);
    4: PCardinal(@Result)^ := GetEnumValue(TypeInfo(T), x);
  end;
end;

class function TEnumHelper<T>.Format(x: T): String;
begin
  case Sizeof(T) of
    1: Result := GetEnumName(TypeInfo(T), PByte(@x)^);
    2: Result := GetEnumName(TypeInfo(T), PWord(@x)^);
    4: Result := GetEnumName(TypeInfo(T), PCardinal(@x)^);
  end;
end;

end.

object Form1: TForm1
  Left = 0
  Top = 13
  Caption = #1051#1072#1073'.'#8470'73. '#1044#1072#1090#1095#1080#1082#1080' '#1082#1080#1089#1083#1086#1088#1086#1076#1072'. '#1043#1088#1072#1092#1080#1082'.'
  ClientHeight = 505
  ClientWidth = 691
  Color = clWindow
  Font.Charset = DEFAULT_CHARSET
  Font.Color = clWindowText
  Font.Height = -11
  Font.Name = 'Tahoma'
  Font.Style = []
  Menu = MainMenu1
  OldCreateOrder = False
  Position = poDesigned
  OnActivate = FormActivate
  OnCreate = FormCreate
  PixelsPerInch = 96
  TextHeight = 13
  object Panel1: TPanel
    Left = 0
    Top = 500
    Width = 691
    Height = 5
    Align = alBottom
    BevelOuter = bvNone
    TabOrder = 0
  end
  object Panel2: TPanel
    Left = 0
    Top = 0
    Width = 691
    Height = 5
    Align = alTop
    BevelOuter = bvNone
    TabOrder = 1
  end
  object Panel3: TPanel
    Left = 0
    Top = 5
    Width = 5
    Height = 495
    Align = alLeft
    BevelOuter = bvNone
    TabOrder = 2
  end
  object Panel4: TPanel
    Left = 686
    Top = 5
    Width = 5
    Height = 495
    Align = alRight
    BevelOuter = bvNone
    TabOrder = 3
  end
  object Panel5: TPanel
    Left = 5
    Top = 5
    Width = 5
    Height = 495
    Align = alLeft
    BevelOuter = bvNone
    TabOrder = 4
  end
  object Chart1: TChart
    Left = 225
    Top = 5
    Width = 461
    Height = 495
    Legend.Title.Visible = False
    Legend.Visible = False
    MarginRight = 80
    MarginTop = 0
    MarginUnits = muPixels
    Title.VertMargin = 0
    View3D = False
    Align = alClient
    BevelOuter = bvNone
    Color = clWindow
    TabOrder = 5
    DefaultCanvas = 'TGDIPlusCanvas'
    ColorPaletteIndex = 13
  end
  object Panel6: TPanel
    Left = 10
    Top = 5
    Width = 215
    Height = 495
    Align = alLeft
    BevelOuter = bvNone
    TabOrder = 6
    object RadioGroup1: TRadioGroup
      Left = 0
      Top = 0
      Width = 215
      Height = 41
      Align = alTop
      Columns = 2
      ItemIndex = 0
      Items.Strings = (
        #1058#1077#1082#1091#1097#1080#1081
        #1057#1086#1093#1088#1072#1085#1105#1085#1085#1099#1081)
      TabOrder = 0
      OnClick = RadioGroup1Click
    end
    object Panel7: TPanel
      Left = 0
      Top = 41
      Width = 215
      Height = 5
      Align = alTop
      BevelOuter = bvNone
      TabOrder = 1
    end
    object ListView1: TListView
      Left = 0
      Top = 46
      Width = 215
      Height = 449
      Align = alClient
      Checkboxes = True
      Columns = <
        item
        end
        item
        end
        item
        end>
      Font.Charset = DEFAULT_CHARSET
      Font.Color = clWindowText
      Font.Height = -13
      Font.Name = 'Tahoma'
      Font.Style = []
      RowSelect = True
      ParentFont = False
      TabOrder = 2
      ViewStyle = vsReport
      OnCustomDrawSubItem = ListView1CustomDrawSubItem
      OnSelectItem = ListView1SelectItem
      OnItemChecked = ListView1ItemChecked
    end
  end
  object MainMenu1: TMainMenu
    Left = 33
    Top = 93
    object N1: TMenuItem
      Caption = #1054#1095#1080#1089#1090#1080#1090#1100
      object N3: TMenuItem
        Caption = #1058#1077#1082#1091#1097#1080#1081
        OnClick = N3Click
      end
      object N2: TMenuItem
        Caption = #1057#1086#1093#1088#1072#1085#1105#1085#1085#1099#1081
        OnClick = N2Click
      end
    end
  end
end

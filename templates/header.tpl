Cookie            : {{.Cookie }}
DataOffset        : {{.DataOffset}}
TableOffset       : {{.TableOffset}}
HeaderVersion     : {{.HeaderVersion}}
MaxTableEntries   : {{.MaxTableEntries}}
BlockSize         : {{.BlockSize}} bytes
CheckSum          : {{.CheckSum}}
ParentUniqueID    : {{.ParentUniqueID}}
ParentTimeStamp   : {{.ParentTimeStamp | printf "%v"}}
Reserved          : {{.Reserved}}
ParentPath        : {{.ParentPath}}
{{range .ParentLocators}}
  PlatformCode               : {{.PlatformCode}}
  PlatformDataSpace          : {{.PlatformDataSpace}}
  PlatformDataLength         : {{.PlatformDataLength}}
  Reserved                   : {{.Reserved}}
  PlatformDataOffset         : {{.PlatformDataOffset}}
  PlatformSpecificFileLocator: {{.PlatformSpecificFileLocator}}
{{end}}

-- Hex dump --

{{.RawData | dump }}
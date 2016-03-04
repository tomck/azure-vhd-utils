Cookie            : {{.Cookie }}
Features          : {{.Features}}
FileFormatVersion : {{.FileFormatVersion}}
HeaderOffset      : {{.HeaderOffset}}
TimeStamp         : {{.TimeStamp | printf "%v" }}
CreatorApplication: {{.CreatorApplication}}
CreatorVersion    : {{.CreatorVersion}}
CreatorHostOsType : {{.CreatorHostOsType}}
PhysicalSize      : {{.PhysicalSize}} bytes
VirtualSize       : {{.VirtualSize}} bytes
DiskGeometry      : {{.DiskGeometry}}
DiskType          : {{.DiskType}}
CheckSum          : {{.CheckSum}}
UniqueID          : {{.UniqueID}}
SavedState        : {{.SavedState | printf "%v" }}

-- Hex dump --

{{.RawData | dump }}

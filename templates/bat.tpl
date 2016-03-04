{{range $index, $value := .}} BAT[{{adj $index}}] : {{$value | printf "0x%X"}}
{{end}}
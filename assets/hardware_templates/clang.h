# block wrapped by "#defkey" and "#end" is a golang text/template.
# "\\n" is an escape for carrior return

#defkey CLANG_HTTP_POST_HEADER
#define HTTP_POST_HEADER \\\n
"POST /channels/devices/%s/upload?
{{ range $i, $e := .Tags }}
{{ if $i }}&{{ end }}{{ $e }}=%s
{{ end }} HTTP/1.1\r\n" \\\n
"Host: %s:8081\r\n" \\\n
"Content-Type: application/json\r\n" \\\n
"AccessToken: abcdefg\r\n" \\\n
"Content-Length: %d\r\n" \\\n
"\r\n"\\n\\n
#end

#defkey CLANG_HTTP_POST_BODY
#define HTTP_POST_BODY \\\n
"{{"{"}}
{{ range $key, $value := .Fields }}
{{if eq $value "int"}}{{ $key }}=%d,
{{else if eq $value "float"}}{{ $key }}=%f,
{{else if eq $value "string"}}{{ $key }}=%s,
{{else if eq $value "boolean"}}{{ $key }}=%s,
{{ end }}
{{ end }}{{"}"}}\r\n"\\n\\n
#end

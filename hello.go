package hello

import (
	"net/http"

	"html/template"
)

func init() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/omikuji", omikujiHandler)
}

var rootTemplate = template.Must(template.New("root").Parse(`
<!Doctype HTML>
<html>
	<head>
		<meta charset=utf-8>
		<title></title>
	</head>
	<body>
		<a href="/omikuji">おみくじ</a>
	</body>
</html>
`))

func rootHandler(w http.ResponseWriter, r *http.Request) {
	rootTemplate.Execute(w, nil)
}

var omikujiTemplate = template.Must(template.New("omikuji").Parse(`
<!Doctype HTML>
<html>
	<head>
		<meta charset=utf-8>
		<title></title>
	</head>
	<body>
		 大吉です!
	</body>
</html>
`))

func omikujiHandler(w http.ResponseWriter, r *http.Request) {
	omikujiTemplate.Execute(w, nil)
}

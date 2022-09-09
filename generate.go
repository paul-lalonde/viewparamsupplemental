package main

import (
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

var simpletemplate = `<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>{{.Title}}</title>
	</head>
	<body>
<ul> {{range .}}
	<li>Viewpoint {{.Viewpoint}} Algorithms breakdown:
		<ul>{{range .Algorithms}}
			<li>	{{.Algorithm}}
				<ul>{{range .Data}}	
					<li>	{{.Resolution}} {{.Datatype}}	<a href="file:{{.Dataset}}/{{.Filename}}"><img style="vertical-align:middle" width=100 height=100 src="file:{{.Dataset}}/{{.Filename}}"></a></li>
				{{end}}</ul>
			</li>
		{{end}}</ul>
	</li>
{{end}}</ul>
	</body>
</html>
`

var flatDirs = []string{"GreekVillaViews", "BistroInteriorViews"}
var hierarchicalDirs = []string{"BistroExteriorViews"}

var CouldNotParse = errors.New("Could not parse")

func ls(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	ret := []string{}
	for _, file := range files {
		ret = append(ret, file.Name())
	}
	return ret
}

type Datum struct {
	Dataset    string
	Viewpoint  string
	Algorithm  string
	Resolution string
	Datatype   string
	Extension  string
	Filename   string
}

type AlgoSet struct {
	Algorithm string
	Data      []Datum
}

type Experiment struct {
	Viewpoint  string
	Algorithms map[string]AlgoSet
	Data       []Datum
}

func extractData(file, dataset string) (exp Datum, err error) {
	exp.Filename = file
	coords_regex := regexp.MustCompile(`\((.*)\)_(.*)`)
	matches := coords_regex.FindStringSubmatch(file)

	if len(matches) < 2 {
		return exp, CouldNotParse
	}
	exp.Viewpoint = strings.ReplaceAll(matches[1], " ", "")
	remainder := matches[2]
	if strings.HasPrefix(remainder, "VIEW_DEPENDENT") {
		remainder = strings.Replace(remainder, "VIEW_DEPENDENT", "VIEWDEPENDENT", 1)
	}

	remain_regex := regexp.MustCompile(`(.*)_(.*)_(.*)\.(.*)`)
	matches = remain_regex.FindStringSubmatch(remainder)

	if len(matches) < 4 {
		return exp, CouldNotParse
	}

	exp.Dataset = dataset
	exp.Algorithm = matches[1]
	exp.Resolution = matches[2]
	exp.Datatype = matches[3]
	exp.Extension = matches[4]

	if exp.Algorithm == "VIEWDEPENDENT" {
		exp.Algorithm = "VIEW_DEPENDENT"
	}

	return exp, err
}

func main() {
	var data []Datum
	for _, dir := range flatDirs {
		files := ls(dir)
		for _, file := range files {
			exp, err := extractData(file, dir)
			if err != nil {
				//log.Println(err, file, exp)
			} else {
				data = append(data, exp)
			}
		}
	}

	// Group the comparisons per viewpoint
	experiments := map[string]Experiment{}
	for _, datum := range data {
		exp, ok := experiments[datum.Viewpoint]
		if !ok {
			exp = Experiment{Viewpoint: datum.Viewpoint, Data: []Datum{}, Algorithms: map[string]AlgoSet{}}
		}
		exp.Data = append(exp.Data, datum)
		experiments[datum.Viewpoint] = exp // TODO(PAL): This is sloppy.  Lots of garbage as I make copies of the experiments to update here.
	}

	for _, exp := range experiments {
		// Break experiments into algorithm views
		for _, datum := range exp.Data {
			algo, ok := exp.Algorithms[datum.Algorithm]
			if !ok {
				algo = AlgoSet{Algorithm: datum.Algorithm, Data: []Datum{}}
			}
			algo.Data = append(algo.Data, datum)
			exp.Algorithms[datum.Algorithm] = algo
		}
	}

	tmpl, err := template.New("test").Parse(simpletemplate)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(os.Stdout, experiments)
	if err != nil {
		panic(err)
	}

	outfile, err := os.OpenFile("supplemental.html", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(outfile, experiments)
	if err != nil {
		panic(err)
	}
}

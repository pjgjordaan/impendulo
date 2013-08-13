package webserver

import(
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/pmd"
	"github.com/godfried/impendulo/tool/checkstyle"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
)

func loadProjectGraphData() (jsonData []map[string]interface{}, err error) {
	projects, err := db.GetProjects(nil)
	if err != nil{
		return
	}
	jsonData = make([]map[string]interface{}, 0)
	for _, p := range projects{
		var subs []*project.Submission
		subs, err = db.GetSubmissions(bson.M{project.PROJECT_ID: p.Id}, bson.M{project.TIME: 1})
		if err != nil{
			return
		}
		if len(subs) == 0{
			continue
		}
		var cur map[string]interface{}
		cur, err = calcData(p.Name, subs)
		if err != nil{
			return
		}
		jsonData = append(jsonData, cur)
	}
	return
}

func loadUserGraphData() (jsonData []map[string]interface{}, err error) {
	users, err := db.GetUsers(nil)
	if err != nil{
		return
	}
	jsonData = make([]map[string]interface{}, 0)
	for _, u := range users{
		var subs []*project.Submission
		subs, err = db.GetSubmissions(bson.M{project.USER: u.Name}, bson.M{project.TIME: 1})
		if err != nil{
			return
		}
		if len(subs) == 0{
			continue
		}
		var cur map[string]interface{}
		cur, err = calcData(u.Name, subs)
		if err != nil{
			return
		}
		jsonData = append(jsonData, cur)
	}
	return
}


func loadResultGraphData(result, name string, files []*project.File) ([]map[string]interface{}, error) {
	var jsonData []map[string]interface{}
	switch result{
	case javac.NAME:
		jsonData = loadJavacGraphData(name, files)
	case jpf.NAME:
	case findbugs.NAME:
	case pmd.NAME:
	case checkstyle.NAME:
	case tool.CODE:
	case tool.SUMMARY:
	default:
		jsonData = loadJUnitGraphData(result, name, files)
	}
	return jsonData, nil
}

func loadJavacGraphData(fname string, files []*project.File) ([]map[string]interface{}) {
	jsonData := make([]map[string]interface{}, 1)
	dataArray := make([]map[string] int64, 0)
	for _, f := range files{
		result, err := db.GetJavacResult(
			bson.M{project.FILEID: f.Id}, nil)
		if err != nil{
			continue
		}
		var y int64
		if result.Success(){
			y = 1
		} else{
			y = 0
		}
		dataArray = append(dataArray, 
			map[string] int64{"x": result.Time/1000, "y": y})
	}
	cur := make(map[string]interface{})
	cur["name"] = fname
	cur["data"] = dataArray
	jsonData[0] = cur
	return jsonData
}

func loadJUnitGraphData(testName, fname string, files []*project.File) ([]map[string]interface{}) {
	jsonData := make([]map[string]interface{}, 1)
	if _, err := db.GetJUnitResult(bson.M{project.NAME: testName}, nil); err != nil{
		return jsonData
	}
	dataArray := make([]map[string] int64, 0)
	for _, f := range files{
		result, err := db.GetJUnitResult(
			bson.M{project.FILEID: f.Id, project.NAME: testName}, nil)
		if err != nil{
			continue
		}
		y := result.Data.Tests - result.Data.Failures - result.Data.Errors  
		dataArray = append(dataArray, map[string] int64{
			"x": result.Time/1000, "y": int64(y)})
	}
	cur := make(map[string]interface{})
	cur["name"] = fname
	cur["data"] = dataArray
	jsonData[0] = cur
	return jsonData
}


func calcData(name string, subs []*project.Submission) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	data["name"] = name
	dataVals := make(map[int64] int64)
	for _, s := range subs{
		v := ((s.Time/1000)/86400)*86400
		dataVals[v] += 1
	}
	dataArray := make([]map[string] int64, len(dataVals))
	index := 0
	for k, v := range dataVals{
		dataArray[index] = map[string] int64{"x": k, "y": v}
		index += 1
	}
	data["data"] = dataArray
	return
}

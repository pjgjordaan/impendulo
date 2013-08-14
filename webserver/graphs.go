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
	"fmt"
	"math"
)

func init(){fmt.Sprint()}

type GraphArgs map[string]interface{}

func loadResultGraphData(result string, files []*project.File) (graphArgs GraphArgs) {
	switch result{
	case javac.NAME:
		graphArgs = loadJavacGraphData(files)
	case jpf.NAME:
		graphArgs = loadJPFGraphData(files)
	case findbugs.NAME:
		graphArgs = loadFindbugsGraphData(files)
	case pmd.NAME:
		graphArgs = loadPMDGraphData(files)
	case checkstyle.NAME:
		graphArgs = loadCheckstyleGraphData(files)
	case "All":
		graphArgs = loadAllGraphData(files)
	case tool.CODE:
	case tool.SUMMARY:
	default:
		graphArgs = loadJUnitGraphData(result, files)
	}
	if graphArgs == nil{
		return
	}
	graphArgs["height"] = 400;
	graphArgs["width"] = 700;
	graphArgs["interpolation"] = "linear";
	return
}

func loadJavacGraphData(files []*project.File) (GraphArgs) {
	graphData := make([]map[string]interface{}, 1)
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
	cur["name"] = "Compilation Success"
	cur["data"] = dataArray
	graphData[0] = cur
	graphArgs := make(map[string]interface{})
	graphArgs["max"] = 1.05
	graphArgs["min"] = -0.05
	graphArgs["renderer"] = "scatterplot"
	graphArgs["series"] = graphData
	graphArgs["yformat"] = map[string]interface{}{"0": "No Compile", "1": "Compiled"}
	return graphArgs
}

func loadJUnitGraphData(testName string, files []*project.File) (GraphArgs) {
	graphArgs := make(map[string]interface{})
	if !db.Contains(db.RESULTS, bson.M{project.NAME: testName}){
		return graphArgs
	}
	graphData := make([]map[string]interface{}, 3)
	max := 0.0
	for _, f := range files{
		result, err := db.GetJUnitResult(
			bson.M{project.FILEID: f.Id, project.NAME: testName}, nil)
		if err != nil{
			continue
		}  
		max = result.AddGraphData(max, graphData)
	}
	graphArgs["max"] = max + max*0.05
	graphArgs["min"] = -max*0.05
	graphArgs["renderer"] = "line"
	graphArgs["series"] = graphData
	return graphArgs
}

func loadJPFGraphData(files []*project.File) (GraphArgs) {
	graphArgs := make(map[string]interface{})
	graphData := make([]map[string]interface{}, 3)
	max := 0.0
	for _, f := range files{
		result, err := db.GetJPFResult(
			bson.M{project.FILEID: f.Id}, nil)
		if err != nil{
			continue
		}  
		max = result.AddGraphData(max, graphData)
	}
	graphArgs["max"] = max + max*0.05
	graphArgs["min"] = -max*0.05
	graphArgs["renderer"] = "line"
	graphArgs["series"] = graphData
	return graphArgs
}

func loadFindbugsGraphData(files []*project.File) (GraphArgs) {
	graphArgs := make(map[string]interface{})
	graphData := make([]map[string]interface{}, 4)
	max := 0.0
	for _, f := range files{
		result, err := db.GetFindbugsResult(
			bson.M{project.FILEID: f.Id}, nil)
		if err != nil || result.Data == nil {
			continue
		}  
		max = result.AddGraphData(max, graphData) 
	}
	graphArgs["max"] = max + max*0.05
	graphArgs["min"] = -max*0.05
	graphArgs["renderer"] = "line"
	graphArgs["series"] = graphData
	return graphArgs
}


func loadCheckstyleGraphData(files []*project.File) (GraphArgs) {
	graphArgs := make(map[string]interface{})
	graphData := make([]map[string]interface{}, 1)
	max := 0.0
	for _, f := range files{
		result, err := db.GetCheckstyleResult(
			bson.M{project.FILEID: f.Id}, nil)
		if err != nil || result.Data == nil {
			continue
		}  
		max = result.AddGraphData(max, graphData)
	}
	graphArgs["max"] = max + max*0.05
	graphArgs["min"] = -max*0.05
	graphArgs["renderer"] = "line"
	graphArgs["series"] = graphData
	return graphArgs
}

func loadPMDGraphData(files []*project.File) (GraphArgs) {
	graphArgs := make(map[string]interface{})
	graphData := make([]map[string]interface{}, 1)
	max := 0.0
	for _, f := range files{
		result, err := db.GetPMDResult(
			bson.M{project.FILEID: f.Id}, nil)
		if err != nil || result.Data == nil {
			continue
		}  
		max = result.AddGraphData(max, graphData) 
	}
	graphArgs["max"] = max + max*0.05
	graphArgs["min"] = -max*0.05
	graphArgs["renderer"] = "line"
	graphArgs["series"] = graphData
	return graphArgs
}

func loadAllGraphData(files []*project.File) (GraphArgs) {
	graphArgs := make(map[string]interface{})
	graphData := make(map[string][]map[string]interface{})
	allMax := make(map[string]float64)
	for _, f := range files{
		results, err := db.GetGraphResults(f.Id)
		if err != nil || results == nil {
			continue
		}  
		for _, result := range results{
			if _, ok := graphData[result.GetName()]; !ok{
				graphData[result.GetName()] = make([]map[string]interface{}, 4)
			}
			allMax[result.GetName()] = result.AddGraphData(allMax[result.GetName()], graphData[result.GetName()])
		} 
	}
	max := 0.0
	for _, v := range allMax{
		max = math.Max(max, v)
	}
	allData := make([]map[string]interface{}, 0)
	for key, val := range graphData{
		scale := max/allMax[key]
		for _, data := range val{
			if data == nil{
				break
			}
			point := data["data"].([]map[string]float64)
			for _, vals := range point{
				vals["y"] *= scale
			}
			data["data"] = point
			allData = append(allData, data)
		} 
	}
	graphArgs["max"] = max + max*0.05
	graphArgs["min"] = -max*0.05
	graphArgs["renderer"] = "line"
	graphArgs["series"] = allData
	return graphArgs
}
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
	"math"
)

//GraphArgs represents arguments which are passed to rickshaw
//in order to draw a graph.
type GraphArgs map[string]interface{}
//GraphData represents the x and y values used to draw the graphs.
type GraphData []map[string]interface{}

//LoadResultGraphData calculates GraphArgs for a given result.
func LoadResultGraphData(result string, files []*project.File) (graphArgs GraphArgs) {
	var graphData GraphData
	max := -1.0
	switch result{
	case javac.NAME:
		graphData, max = loadJavacGraphData(files)
	case jpf.NAME:
		graphData, max = loadJPFGraphData(files)
	case findbugs.NAME:
		graphData, max = loadFindbugsGraphData(files)
	case pmd.NAME:
		graphData, max = loadPMDGraphData(files)
	case checkstyle.NAME:
		graphData, max = loadCheckstyleGraphData(files)
	case "All":
		graphData, max = loadAllGraphData(files)
	case tool.CODE:
	case tool.SUMMARY:
	default:
		graphData, max = loadJUnitGraphData(result, files)
	}
	if max == -1.0{
		return
	}
	graphArgs = make(GraphArgs)
	graphArgs["max"] = max+max*0.05
	graphArgs["series"] = graphData
	if result == javac.NAME{
		graphArgs["yformat"] = map[string]interface{}{"0": "Failure", "1": "Success"}
		graphArgs["min"]= -0.05
	}
	graphArgs["height"] = 400
	graphArgs["width"] = 700
	graphArgs["interpolation"] = "linear"
	graphArgs["renderer"] = "line"
	return
}

func loadJavacGraphData(files []*project.File) (GraphData, float64) {
	graphData := make(GraphData, 1)
	for _, f := range files{
		result, err := db.GetJavacResult(
			bson.M{project.FILEID: f.Id}, nil)
		if err != nil{
			continue
		}
		result.AddGraphData(0, graphData)
	}
	return graphData, 1
}

func loadJUnitGraphData(testName string, files []*project.File) (GraphData, float64) {
	graphData := make(GraphData, 3)
	max := -1.0
	for _, f := range files{
		result, err := db.GetJUnitResult(
			bson.M{project.FILEID: f.Id, project.NAME: testName}, nil)
		if err != nil{
			continue
		}  
		max = result.AddGraphData(max, graphData)
	}
	return graphData, max
}

func loadJPFGraphData(files []*project.File) (GraphData, float64) {
	graphData := make(GraphData, 3)
	max := -1.0
	for _, f := range files{
		result, err := db.GetJPFResult(
			bson.M{project.FILEID: f.Id}, nil)
		if err != nil{
			continue
		}  
		max = result.AddGraphData(max, graphData)
	}
	return graphData, max
}

func loadFindbugsGraphData(files []*project.File) (GraphData, float64) {
	graphData := make(GraphData, 4)
	max := -1.0
	for _, f := range files{
		result, err := db.GetFindbugsResult(
			bson.M{project.FILEID: f.Id}, nil)
		if err != nil || result.Data == nil {
			continue
		}  
		max = result.AddGraphData(max, graphData) 
	}
	return graphData, max
}


func loadCheckstyleGraphData(files []*project.File) (GraphData, float64) {
	graphData := make(GraphData, 1)
	max := -1.0
	for _, f := range files{
		result, err := db.GetCheckstyleResult(
			bson.M{project.FILEID: f.Id}, nil)
		if err != nil || result.Data == nil {
			continue
		}  
		max = result.AddGraphData(max, graphData)
	}
	return graphData, max
}

func loadPMDGraphData(files []*project.File) (GraphData, float64) {
	graphData := make(GraphData, 1)
	max := -1.0
	for _, f := range files{
		result, err := db.GetPMDResult(
			bson.M{project.FILEID: f.Id}, nil)
		if err != nil || result.Data == nil {
			continue
		}  
		max = result.AddGraphData(max, graphData) 
	}
	return graphData, max
}

func loadAllGraphData(files []*project.File) (GraphData, float64) {
	graphData := make(map[string]GraphData)
	allMax := make(map[string]float64)
	for _, f := range files{
		results, err := db.GetGraphResults(f.Id)
		if err != nil || results == nil {
			continue
		}  
		for _, result := range results{
			if _, ok := graphData[result.GetName()]; !ok{
				graphData[result.GetName()] = make(GraphData, 4)
			}
			allMax[result.GetName()] = result.AddGraphData(allMax[result.GetName()], graphData[result.GetName()])
		} 
	}
	max := 0.0
	for _, v := range allMax{
		max = math.Max(max, v)
	}
	allData := make(GraphData, 0)
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
	return allData, max
}
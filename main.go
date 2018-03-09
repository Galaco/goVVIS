package main

import (
	"github.com/galaco/bsp"
	"github.com/galaco/bsp/lumps/datatypes/node"
	"github.com/galaco/bsp/lumps/datatypes/face"
	"github.com/galaco/vmf"
	"log"
	"os"
	"bytes"
	"fmt"
	"strconv"
	"github.com/galaco/bsp/lumps/datatypes/leaf"
	"strings"
	"path/filepath"
)

var Leafs []leaf.Leaf


func main() {
	fmt.Println("DormantLemon/Galaco's VVIS compiler")
	// Load Arguments
	Args := ParseCmdArguments()

	if Args.LowPriority {
		//SetLowPriority()
	}
	// Load bsp file
	file := ImportBSP(Args.File)

	if len(file.GetLump(bsp.LUMP_NODES).GetContents().GetData().([]node.Node)) == 0 ||
		len(file.GetLump(bsp.LUMP_FACES).GetContents().GetData().([]face.Face)) == 0 {
		log.Fatal("Empty map.")
	}

	entities := ParseEntities(file.GetLump(bsp.LUMP_ENTITIES).GetContents().GetData().(string))

	if Args.UseRadius == false {
		radius := DetermineVisRadius(&entities)
		if radius > 0.0 {
			Args.UseRadius = true
			Args.VisRadius = radius * radius
		}
	}

	if Args.UseRadius {
		Leafs = file.GetLump(bsp.LUMP_LEAFS).GetContents().GetData().([]leaf.Leaf)
		MarkLeavesAsRadial(&Leafs)
	}

	portalFilename := Args.File
	// Does this even do anything except print?
	if Args.TmpIn != "" {
		// sprintf ( portalfile, "%s%s", inbase, argv[i] );
		// Q_StripExtension( portalfile, portalfile, sizeof( portalfile ) );
	}
	portalFilename = strings.TrimSuffix(portalFilename, filepath.Ext(portalFilename))
	portalFilename += ".prt"
	fmt.Println("Reading " + portalFilename)

	LoadPortals (portalFilename)

	if Args.TraceClusterStart < 0 {
		// CalcVis()
		// CalcPAS()

		// BuilfClusterTable()
		// CalcVisibleFogVolumes()
		// CalcDistanceFromLeavesToWater()

		// visdatasize = vismap_p - dvisdata;
		// Msg ("visdatasize:%i  compressed from %i\n", visdatasize, originalvismapsize*2);

		// Msg ("writing %s\n", mapFile);
		// WriteBSPFile (mapFile);
	} else {
		// if ( g_TraceClusterStart < 0 || g_TraceClusterStart >= portalclusters || g_TraceClusterStop < 0 || g_TraceClusterStop >= portalclusters ) {
		//	Error("Invalid cluster trace: %d to %d, valid range is 0 to %d\n", g_TraceClusterStart, g_TraceClusterStop, portalclusters-1 );
		// }
		// CalcVisTrace ();
		// WritePortalTrace(source);
	}
}

// Read and parse BSP created by vbsp
func ImportBSP(filename string) bsp.Bsp {
	file,err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	fi,err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	fileData := make([]byte, fi.Size())
	file.Read(fileData)
	file.Close()

	Reader := bsp.NewReader()
	Reader.SetBuffer(fileData)

	return Reader.Read()
}

func ParseEntities(entdata string) vmf.Node {
	buffer := bytes.NewBufferString(entdata)

	reader := vmf.NewReader(buffer)
	parsed,err := reader.Read()
	if err != nil {
		FatalError("Unable to parse entities")
	}
	return parsed.Unclassified
}

func FatalError(text string) {
	log.Fatal(text)
}

// Determine max leaf Radius from fog controller farz
func DetermineVisRadius(entities *vmf.Node) float64 {
	radius := -1.0
	for _,i := range *entities.GetAllValues() {
		n := i.(vmf.Node)
		if n.GetProperty("classname") == "env_fog_controller" {
			s := n.GetProperty("farz")
			if s == "" {
				return radius
			}
			f,err := strconv.ParseFloat(s, 64)
			if err != nil {
				if f == 0.0 {
					radius = -1.0
				} else {
					radius = f
				}
			}
			break
		}
	}
	return radius
}

//Mark leaves as radius
func MarkLeavesAsRadial(leafs *[]leaf.Leaf) {
	for i,_ := range *leafs {
		(*leafs)[i].Flags |= leaf.LEAF_FLAGS_RADIAL
	}
}

func LoadPortals(portalFilename string) {

}
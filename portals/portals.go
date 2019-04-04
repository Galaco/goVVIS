package portals

import (
	"log"
	"fmt"
	"os"
	"strings"
	"github.com/galaco/bsp/lumps/datatypes/portal"
	"github.com/galaco/bsp/lumps/datatypes/visibility"
	"github.com/galaco/bsp/lumps"
	"github.com/go-gl/mathgl/mgl32"
)

const PORTALFILE = "PRT1"
const MAX_POINTS_ON_WINDING = 64
const MAX_POINTS_ON_FIXED_WINDING = 12

type Vec3 mgl32.Vec3

type VStatus int8
const (
	Stat_none VStatus = iota
	Stat_working VStatus = iota
	Stat_done VStatus = iota
)

type Winding struct {
	Original int32 //qboolean
	NumPoints int32
	Points [MAX_POINTS_ON_FIXED_WINDING]Vec3
}

type Vis struct {
	NumClusters int32
	Bitofs [8][2]int32	// bitofs[numclusters][2]
}

type Plane struct {
	Normal Vec3
	Dist float32
}

type Portal struct {
	Plane Plane
	Leaf int32
	Origin Vec3
	Radius float32
	Winding *Winding
	Status VStatus
	PortalFront *byte
	PortalFlood *byte
	PortalVis *byte
	NumMightSee int32
}

type Leaf struct {
	Portals []*Portal
}

type Exports struct {
	PortalLongs int32
	LeafLongs int32
	UncompressedVis []byte
	VismapEnd *[]byte
	PortalClusters int
}

func Load(name string, useMPI bool, visLump *lumps.Visibility) Exports {
	var i int32
	var j int32
	var p *Portal
	var l *Leaf
	var magic [80]byte //char
	var numpoints int32
	var w *Winding
	var leafnums [2]int32
	var plane Plane
	var portalExports Exports
	var dvis Vis

	var f *os.File
	var err error
	var numPortals int32

	// Open the portal file.
	if useMPI {
		/*
		// If we're using MPI, copy off the file to a temporary first. This will download the file
		// from the MPI master, then we get to use nice functions like fscanf on it.
		tempPath [MAX_PATH]byte //char
		tempFile [MAX_PATH]byte //char
		if  GetTempPath( sizeof( tempPath ), tempPath ) == 0 {
			log.Fatalf( "LoadPortals: GetTempPath failed.\n" )
		}

		if GetTempFileName( tempPath, "vvis_portal_", 0, tempFile ) == 0 {
			log.Fatalf( "LoadPortals: GetTempFileName failed.\n" )
		}

		// Read all the data from the network file into memory.
		FileHandle_t hFile = g_pFileSystem->Open(name, "r");
		if hFile == FILESYSTEM_INVALID_HANDLE {
			log.Fatalf( "LoadPortals( %s ): couldn't get file from master.\n", name )
		}

		CUtlVector<char> data;
		data.SetSize( g_pFileSystem->Size( hFile ) );
		g_pFileSystem->Read( data.Base(), data.Count(), hFile );
		g_pFileSystem->Close( hFile );

		// Dump it into a temp file.
		f = fopen( tempFile, "wt" );
		fwrite( data.Base(), 1, data.Count(), f );
		fclose( f );

		// Open the temp file up.
		f = fopen( tempFile, "rSTD" ); // read only, sequential, temporary, delete on close
		*/
	} else {
		f,err = os.Open(name)
	}

	if err != nil {
		log.Fatalf("LoadPortals: couldn't read %s\n",name)
	}

	if n,_ :=fmt.Fscanf(f,"%79s\n%i\n%i\n", &magic, &portalExports.PortalClusters, &numPortals); n != 3 {
		log.Fatalf("LoadPortals %s: failed to read header", name)
	}
	if strings.Compare(string(magic[:]), PORTALFILE) > 0 {
		log.Fatalf("LoadPortals %s: not a portal file", name)
	}

	fmt.Printf("%4i portalclusters\n", portalExports.PortalClusters)
	fmt.Printf("%4i numportals\n", numPortals)

	if numPortals * 2 >= portal.MAX_PORTALS {
		log.Fatalf("The map overflows the max portal count (%d of max %d)!\n", numPortals, portal.MAX_PORTALS / 2 )
	}

	// these counts should take advantage of 64 bit systems automatically
	leafbytes := ((portalExports.PortalClusters + 63) &~ 63) >> 3
	portalExports.LeafLongs = int32(leafbytes / 4) //4 = sizeof int32 | x86 long

	portalbytes := ((numPortals * 2 + 63) &~ 63) >> 3
	portalExports.PortalLongs = portalbytes / 4 //4 = sizeof int32 | x86 long

	// each file portal is split into two memory portals
	//portals := (portal_t*)malloc(2*numPortals*sizeof(portal_t))
	//memset (portals, 0, 2*numPortals*sizeof(portal_t))
	portals := make([]Portal, 2*numPortals)

	//leafs = (leaf_t*)malloc(ortalClusters*int32(unsafe.Sizeof(leaf.Leaf{})))
	//memset (leafs, 0, portalClusters*int32(unsafe.Sizeof(leaf.Leaf{})))
	leafs := make([]Leaf, portalExports.PortalClusters)

	originalvismapsize := portalExports.PortalClusters*leafbytes
	portalExports.UncompressedVis = make([]byte, originalvismapsize)

	vismap := visLump.ToBytes()
	dvis.NumClusters = int32(portalExports.PortalClusters)
	//buf := new(bytes.Buffer)
	//binary.Write(buf, binary.LittleEndian, &dvis.Bitofs[portalClusters])
	//vismap_p := buf.Bytes()

	// need to think about this solution
	// *byte = *byte + int
	portalExports.VismapEnd = (vismap + visibility.MAX_MAP_VISIBILITY)

	pIndex := 0
	for i, p = 0, &portals[pIndex]; i < numPortals; i++ {
		if n,_ :=fmt.Fscanf (f, "%i %i %i ", &numpoints, &leafnums[0], &leafnums[1]); n != 3 {
			log.Fatalf("LoadPortals: reading portal %i", i)
		}
		if numpoints > MAX_POINTS_ON_WINDING {
			log.Fatalf("LoadPortals: portal %i has too many points", i)
		}
		if leafnums[0] > int32(portalExports.PortalClusters) || leafnums[1] > int32(portalExports.PortalClusters) {
			log.Fatalf("LoadPortals: reading portal %i", i)
		}

		p.Winding = newWinding (numpoints)
		w = p.Winding
		w.Original = 1 //true b/c qboolean
		w.NumPoints = numpoints

		for j=0; j<numpoints; j++ {
			v := [3]float32{} //actually a double
			k := 0

			// scanf into double, then assign to vec_t
			// so we don't care what size vec_t is
			if n,_ := fmt.Fscanf (f, "(%lf %lf %lf ) ", &v[0], &v[1], &v[2]); n != 3 {
				log.Fatalf("LoadPortals: reading portal %i", i);
			}
			for k=0; k<3; k++ {
				w.Points[j][k] = v[k]
			}
		}
		fmt.Fscanf (f, "\n")

		// calc plane
		PlaneFromWinding (w, &plane)

		// create forward portal
		l = &leafs[leafnums[0]]
		l.Portals = append(l.Portals, p)

		p.Winding = w

		VectorSubtract (vec3_origin, plane.Normal, p.Plane.Normal)
		p.Plane.Normal = vec3_origin.Sub(plane.Normal)
		p.Plane.Dist = - plane.Dist;
		p.Leaf = leafnums[1]
		SetPortalSphere (p)
		pIndex++
		p = &portals[pIndex]

		// create backwards portal
		l = &leafs[leafnums[1]]
		//l->portals.AddToTail(p)
		l.Portals = append(l.Portals, p)

		p.Winding = newWinding(w.NumPoints)
		p.Winding.NumPoints = w.NumPoints
		for j=0 ; j< w.NumPoints ; j++ {
			VectorCopy (w.Points[w.NumPoints-1-j], p.Winding.Points[j])
		}

		p.Plane = plane
		p.Leaf = leafnums[0]
		SetPortalSphere (p)
		pIndex++
		p = &portals[pIndex]
	}

	//fclose (f)
	f.Close()

	return portalExports
}


func newWinding(points int32) *Winding {
	var w *Winding

	if points > MAX_POINTS_ON_WINDING {
		log.Fatalf("NewWinding: %i points, max %d", points, MAX_POINTS_ON_WINDING)
	}
	// @TODO WTF is this?
	//size = (int)(&((winding_t *)0)->points[points]);
	//w = (winding_t*)malloc (size);
	//memset (w, 0, size);

	return w
}

func SetPortalSphere (p *Portal) {
	var i int32
	var total Vec3
	var dist Vec3
	var w *Winding
	var r float32
	var bestr float32

	w = p.Winding
	VectorCopy (vec3_origin, total)
	for i = 0; i < w.NumPoints; i++ {
		VectorAdd (total, w.Points[i], total)
		//total = total.Add(w.Points[i])
	}

	for i=0 ; i<3 ; i++ {
		total[i] /= w.NumPoints
	}

	bestr = 0;
	for i=0; i<w.NumPoints ; i++ {
		VectorSubtract (w.Points[i], total, dist)
		//dist = w.Points[i].Sub(total)
		r = VectorLength (dist)
		//r = dist.Length()
		if r > bestr {
			bestr = r
		}
	}
	VectorCopy (total, p.Origin)
	p.Radius = bestr
}

func PlaneFromWinding (w *Winding, plane *Plane) {
	var v1 Vec3
	var v2 Vec3

	// calc plane
	VectorSubtract (w.Points[2], w.Points[1], v1);
	//v1 = w.Points[2].Sub(w.Points[1])
	VectorSubtract (w.Points[0], w.Points[1], v2);
	//v2 = w.Points[0].Sub(w.Points[1])
	CrossProduct (v2, v1, plane.Normal);
	//plane.Normal = v2.Cross(v1)
	VectorNormalize (plane.Normal);
	//plane.Normal = plane.Normal.Normalize()
	plane.Dist = DotProduct (w.Points[0], plane.Normal);
	//plane.Dist = w.Points[0].Dot(plane.Normal)
}

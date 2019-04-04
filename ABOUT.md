package main



```
func main() {
start timer

getMapPath()
loadBSP()
assert NumNodes & numFaces > 0
parseEntities()
checkBspForRadiusCulling()
if !radiusCulling {
	determineVisRadius()
	if determined > 0 {
		radiusCulling = true
		visRadius  = determined^2 
	}
}
if radiusCulling {
	markLeavesAsRadial()
}
if inbase[0] == 0 {
	copy portalfile to source
}
LoadPortals(portalfile)

//dont write results when doing a trace
if traceClusterStart < 0 {
	CalcVis()
	CalcPAS()

	//Map from cluster to leaves
	BuildClusterTable();

	//Calc distance from leaves to water
	CalcFogVolumes()
	CalcDistanceFromLeavesToWater()

	WriteBSP()
} else {
	assert valid (traceClusterStart > -1 & < portalClusters, traceClusterStop > -1 & < portalClusters)

	if useMPI {
		warn that cant compile trace in this mode
	}
	CalcVisTrace()
	WritePortalTrace(source)
}

stop timer
write duration
}
```
import { Ion, Viewer, createWorldTerrain, createOsmBuildings, Cartesian3, Cartesian2, HeadingPitchRoll, HeightReference, Math, Transforms, Color, LabelStyle, WebMercatorProjection, HorizontalOrigin, VerticalOrigin, PolylineDashMaterialProperty, ArcType, GeographicProjection, SceneMode, MapMode2D } from "cesium";
import "cesium/Widgets/widgets.css";
import "../src/css/main.css";

function readTextFile(file, callback) {
  var rawFile = new XMLHttpRequest();
  rawFile.overrideMimeType("application/json");
  rawFile.open("GET", file, true);
  rawFile.onreadystatechange = function () {
    if (rawFile.readyState === 4 && rawFile.status == "200") {
      callback(rawFile.responseText);
    }
  }
  rawFile.send(null);
}
var data;

// Your access token can be found at: https://cesium.com/ion/tokens.
// This is the default access token
Ion.defaultAccessToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiJlYWE1OWUxNy1mMWZiLTQzYjYtYTQ0OS1kMWFjYmFkNjc5YzciLCJpZCI6NTc3MzMsImlhdCI6MTYyNzg0NTE4Mn0.XcKpgANiY19MC4bdFUXMVEBToBmqS8kuYpUlxJHYZxk';

// Initialize the Cesium Viewer in the HTML element with the `cesiumContainer` ID.
const viewer = new Viewer('cesiumContainer', {
  infoBox: true,
  terrainProvider: createWorldTerrain(),
  sceneMode : SceneMode.SCENE3D,
  timeline : false,
  animation : false
});

// Add Cesium OSM Buildings, a global 3D buildings layer.
viewer.scene.primitives.add(createOsmBuildings());
viewer.scene.fxaa = false
viewer.resolutionScale = 2

//usage:
readTextFile("out/data_perception.json", function (text) {
  data = JSON.parse(text);
  //console.log(data);
  //parse json entity data
  //var mstdata = JSON.parse("data_mst.txt")
  //console.log(mstdata.entities);

  var heading = Math.toRadians(0);
  var pitch = 0;
  var roll = 0;
  var hpr = new HeadingPitchRoll(heading, pitch, roll);

  for (var i = 0; i < data.entities.length; i++) {
    //console.log(data.entities[i].name);

    var pos = Cartesian3.fromDegrees(
      data.entities[i].pos.Lon,
      data.entities[i].pos.Lat,
      data.entities[i].pos.Alt * 1000 //kilometers to meters 
    );

    var entity = viewer.entities.add({
      position: pos,
      //label: {
      //  //id: data.entities[i].id + "",
      //  //text: data.entities[i].id + "",
      //  //font: '18px sans-serif',
      //  fillColor: Color.BLACK,
      //  outlineColor: Color.BLACK,
      //  outlineWidth: 1.0,
      //  pixelOffset: new Cartesian2(0,-16),
      //  style: LabelStyle.FILL_AND_OUTLINE,
      //  scale : 1
      //},
      id: data.entities[i].id,
      name: data.entities[i].name,
      point: {
        color: Color.LIGHTSKYBLUE,
        pixelSize: 3.0
      }
    });
    //entity.label.show = true;
  }


  for (var i = 0; i < data.entities.length; i++) {
    var max = data.entities[i].percept.length;
    if (max > 0) {
      max = 0;
    }
    for (var j = 0; j < max; j++) {
      //console.log(data.entities[i].percept[j]["Id"] + " -> " + //data.entities[i].percept[j]["Weight"])
      var other_id = data.entities[i].percept[j]["Id"];

      var mid = Cartesian3.fromDegrees(
        (data.entities[i].pos.Lon + data.entities[other_id].pos.Lon)/2,
        (data.entities[i].pos.Lat+data.entities[other_id].pos.Lat)/2,
        (data.entities[i].pos.Alt * 1000 + data.entities[other_id].pos.Alt * 1000)/2 //kilometers to meters 
      );

      var greenLine = viewer.entities.add({
        position: mid,
        //label: {
        //  id: data.entities[i].id + "<-->" + data.entities[other_id].id,
        //  //text: data.entities[i].percept[j]["Wt"] + "", //data.entities[i].percept[j]["Wt"] + "",
        //  font: '0.1px sans-serif',
        //  fillColor: Color.WHITE,
        //  outlineColor: Color.WHITE,
        //  outlineWidth: 1.0,
        //  pixelOffset: new Cartesian2(-20,-20),
        //  style: LabelStyle.FILL_AND_OUTLINE,
        //  scale : 100
        //},
        name:
          data.entities[i].name + " <---> " + data.entities[other_id].name + ": dist=" + data.entities[i].percept[j]["Wt"],
        polyline: {
          positions: Cartesian3.fromDegreesArrayHeights([
            data.entities[other_id].pos.Lon,
            data.entities[other_id].pos.Lat,
            data.entities[other_id].pos.Alt * 1000, //kilometers to meters 
            data.entities[i].pos.Lon,
            data.entities[i].pos.Lat,
            data.entities[i].pos.Alt * 1000 //kilometers to //meters ,
          ]),
          width: 1,
          arcType: ArcType.NONE,
          material: new PolylineDashMaterialProperty({
            color: Color.GREEN,
          })
        },
      });
      //greenLine.label.show = false;
    }
  }
});

readTextFile("out/data_mst.json", function (text) {
  var mst_data = JSON.parse(text);
  //console.log(data);
  for (var i = 0; i < mst_data.entities.length; i++) {
    if (mst_data.entities[i].mst == null) {
      continue
    }
    var max = mst_data.entities[i].mst.length;
    for (var j = 0; j < max; j++) {
      //console.log(mst_data.entities[i].mst[j]["Id"] + " -> " + mst_data.entities[i].mst[j]["Wt"])
      var other_id = mst_data.entities[i].mst[j]["Id"]
      var redLine = viewer.entities.add({
        position: Cartesian3.fromDegrees(
          mst_data.entities[i].pos.Lon,
          mst_data.entities[i].pos.Lat,
          mst_data.entities[i].pos.Alt * 1000 //kilometers to meters 
        ),
        //label: {
        //  id: mst_data.entities[i].id + "<-->" + mst_data.entities[other_id].id,
        //  text: "",
        //},
        name: mst_data.entities[i].name + " To " + mst_data.entities[other_id].name,
        polyline: {
          positions: Cartesian3.fromDegreesArrayHeights([
            mst_data.entities[other_id].pos.Lon,
            mst_data.entities[other_id].pos.Lat,
            mst_data.entities[other_id].pos.Alt * 1000, //kilometers to meters 
            mst_data.entities[i].pos.Lon,
            mst_data.entities[i].pos.Lat,
            mst_data.entities[i].pos.Alt * 1000 //kilometers to meters 
          ]),
          width: 1,
          arcType: ArcType.NONE,
          material: new PolylineDashMaterialProperty({
            color: Color.RED,
          })
        },
      });
      //redLine.label.show = true;
    }
  }
});
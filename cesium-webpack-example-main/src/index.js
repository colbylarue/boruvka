import { Ion, Viewer, createWorldTerrain, createOsmBuildings, Cartesian3, HeadingPitchRoll, Math, Transforms, Color, PolylineDashMaterialProperty, ArcType } from "cesium";
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
//usage:
readTextFile("out/data.json", function (text) {
  data = JSON.parse(text);
  console.log(data);
  // Your access token can be found at: https://cesium.com/ion/tokens.
  // This is the default access token
  Ion.defaultAccessToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiJlYWE1OWUxNy1mMWZiLTQzYjYtYTQ0OS1kMWFjYmFkNjc5YzciLCJpZCI6NTc3MzMsImlhdCI6MTYyNzg0NTE4Mn0.XcKpgANiY19MC4bdFUXMVEBToBmqS8kuYpUlxJHYZxk';

  // Initialize the Cesium Viewer in the HTML element with the `cesiumContainer` ID.
  const viewer = new Viewer('cesiumContainer', {
    infoBox: false,
    terrainProvider: createWorldTerrain()
  });

  // Add Cesium OSM Buildings, a global 3D buildings layer.
  viewer.scene.primitives.add(createOsmBuildings());


  //parse json entity data
  //var mstdata = JSON.parse(minst)
  //console.log(data.entities);

  var heading = Math.toRadians(0);
  var pitch = 0;
  var roll = 0;

  for (var i = 0; i < data.entities.length; i++) {
    console.log(data.entities[i].name);

    var pos = Cartesian3.fromDegrees(
      data.entities[i].position.Longitude * 180 / Math.PI,
      data.entities[i].position.Latitude * 180 / Math.PI,
      data.entities[i].position.Altitude * 1000 //kilometers to meters 
    );

    var hpr = new HeadingPitchRoll(heading, pitch, roll);
    var or = Transforms.headingPitchRollQuaternion(
      pos,
      hpr
    );

    viewer.entities.add({
      name: data.entities[i].name,
      position: pos,
      point: {
        color: Color.LIGHTSKYBLUE,
        pixelSize: 5
      }
    });

    // get only the first 8 connected sats for optimization reasons
    // this should be a sorted list by value of distance. 
    var max = data.entities[i].perception.length;
    //if (max > 8) {
    //  max = 8;
    //}
    for (var j = 0; j < max; j++) {
      //console.log(data.entities[i].perception[j]["Id"] + " -> " + //data.entities[i].perception[j]["Weight"])
      var other_id = data.entities[i].perception[j]["Id"]
      const greenLine = viewer.entities.add({
        name:
          data.entities[i].name + " To " + data.entities[other_id].name,
        polyline: {
          positions: Cartesian3.fromDegreesArrayHeights([
            data.entities[other_id].position.Longitude * 180 / Math.PI,
            data.entities[other_id].position.Latitude * 180 / Math.PI,
            data.entities[other_id].position.Altitude * 1000, //kilometers to meters 
            data.entities[i].position.Longitude * 180 / Math.PI,
            data.entities[i].position.Latitude * 180 / Math.PI,
            data.entities[i].position.Altitude * 1000 //kilometers to //meters ,
          ]),
          width: 1,
          arcType: ArcType.NONE,
          material: new PolylineDashMaterialProperty({
            color: Color.GREEN,
          })
        },
      });
    }
    if ( data.entities[i].mst == null ){
      continue
    }
    var max = data.entities[i].mst.length;
    for (var j = 0; j < max; j++) {
      //console.log(mstdata.entities[i].mst[j]["Id"] + " -> " + mstdata.entities[i].mst[j]["Weight"])
      var other_id = data.entities[i].mst[j]["Id"]
      const redLine = viewer.entities.add({
        name:
        data.entities[i].name + " To " + data.entities[other_id].name,
        polyline: {
          positions: Cartesian3.fromDegreesArrayHeights([
            data.entities[other_id].position.Longitude * 180 / Math.PI,
            data.entities[other_id].position.Latitude * 180 / Math.PI,
            data.entities[other_id].position.Altitude * 1000, //kilometers to meters 
            data.entities[i].position.Longitude * 180 / Math.PI,
            data.entities[i].position.Latitude * 180 / Math.PI,
            data.entities[i].position.Altitude * 1000 //kilometers to meters ,
          ]),
          width: 1,
          arcType: ArcType.NONE,
          material: new PolylineDashMaterialProperty({
            color: Color.RED,
          })
        },
      });
    }
  }
});

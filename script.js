window.onload = function() {
  var w = window.innerWidth;
  var h = window.innerHeight;
  var url = "https://w.soundcloud.com/player/?url=https%3A//api.soundcloud.com/tracks/97542154&amp;auto_play=false&amp;hide_related=false&amp;show_comments=true&amp;show_user=true&amp;show_reposts=false&amp;visual=true";
  function iframe(url) {
    return "<iframe width='100%' height='200' scrolling='no' frameborder='no' src='" + url + "'></iframe>"
  }

  $(document).on("mouseout", ".d3-tip", function() {
    tip.hide()
  });

  var tip = d3.tip()
  .attr('class', 'd3-tip')
  .offset([0, 0])
  .direction("e")
  .html(function(d) {
    return iframe(d.properties.url);
  });

  //Define map projection
  var projection = d3.geo.albersUsa()
  // .translate([w/2, h/2])
  .translate([360, 200])
  .scale([800]);

  //Define path generator
  var path = d3.geo.path()
  .projection(projection);

  //Define quantize scale to sort data values into buckets of color
  var color = d3.scale.quantize()
  .range(["rgb(237,248,233)","rgb(186,228,179)","rgb(116,196,118)","rgb(49,163,84)","rgb(0,109,44)"]);
  //Colors taken from colorbrewer.js, included in the D3 download

  //Create SVG element
  var svg = d3.select("body")
  .append("svg")
  .attr("width", 720)
  .attr("height", 400);

  svg.call(tip);

  //Load in agriculture data
  d3.csv("us-ag-productivity-2004.csv", function(data) {

    //Set input domain for color scale
    color.domain([
      d3.min(data, function(d) { return d.value; }),
      d3.max(data, function(d) { return d.value; })
    ]);

    //Load in GeoJSON data
    d3.json("us-states.json", function(json) {

      //Merge the ag. data and GeoJSON
      //Loop through once for each ag. data value
      for (var i = 0; i < data.length; i++) {

        //Grab state name
        var dataState = data[i].state;

        //Grab data value, and convert from string to float
        var dataValue = parseFloat(data[i].value);

        //Find the corresponding state inside the GeoJSON
        for (var j = 0; j < json.features.length; j++) {

          var jsonState = json.features[j].properties.name;

          if (dataState == jsonState) {

            //Copy the data value into the JSON
            var state = json.features[j];
            state.properties.value = dataValue;
            state.properties.url = url;

            //Stop looking through the JSON
            break;

          }
        }
      }

      //Bind data and create one path per GeoJSON feature
      svg.selectAll("path")
      .data(json.features)
      .enter()
      .append("path")
      .attr("d", path)
      .style("fill", function(d) {
        //Get data value
        var value = d.properties.value;

        if (value) {
          //If value exists…
          return color(value);
        } else {
          //If value is undefined…
          return "#ccc";
        }
      })
      // try embedding sc widgets in d3.tip a la http://bl.ocks.org/Caged/6476579
      // try to get them to play and all that nice stuff.
      // .on('mouseover', tip.show)
      // .on('mouseout', tip.hide)
      .on("mouseover", function(d) {
        tip.show(d)
      })
      .on("mouseout", function(d) {
        if ($(".d3-tip").is(":hover")) {
          return;
        } else {
          tip.hide(d)
          return;
        }
      });

    });

  });

}

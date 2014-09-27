window.onload = function() {
  var w = window.innerWidth;
  var h = window.innerHeight;
  function iframe(url) {
    return "<iframe width='100%' height='200' scrolling='no' frameborder='no' src='" + url + "'></iframe>"
  }
  function img(imgUrl, linkUrl) {
    return "<img class='album-art' src='" + imgUrl + "' data-url='" + linkUrl + "' />";
  }

  $(document).on("mouseout", ".d3-tip", function() {
    tip.hide()
  });
  $(document).on("click", ".album-art", function() {
    SC.oEmbed(this.dataset.url, {auto_play: true}, function(oembed){
      var container = document.querySelector(".play-container");
      container.innerHTML = oembed.html;
    });
  });

  var tip = d3.tip()
  .attr('class', 'd3-tip')
  .offset([0, -10])
  .direction("e")
  .html(function(d) {
    var p = d.properties;
    if (p.artwork) {
      return img(p.artwork, p.link) + "<div>" + p.title + ": " + p.playcount + " plays.</div>";
    } else {
      return "<div>No data found</div>";
    }
  });

  var projection = d3.geo.albersUsa()
  .translate([360, 200])
  .scale([800]);

  var path = d3.geo.path()
  .projection(projection);

  var svg = d3.select(".svg-container")
  .append("svg")
  .attr("width", 720)
  .attr("height", 400);

  svg.call(tip);

  d3.csv("random_state_data.csv", function(data) {

    var colorScale = ["rgb(237,248,233)","rgb(186,228,179)","rgb(116,196,118)","rgb(49,163,84)","rgb(0,109,44)"];

    var color = d3.scale.quantize()
    .range(colorScale).domain([
      d3.min(data, function(d) { return +d.count; }),
      d3.max(data, function(d) { return +d.count; })
    ]);

    d3.select(".legend-container")
    .selectAll(".palette")
    .data([colorScale])
    .enter().append("span")
    .attr("class", "palette")
    .attr("title", function(d) { return d.key; })
    .selectAll(".swatch")
    .data(function(d) { return colorScale; })
    .enter().append("span")
    .attr("class", "swatch")
    .style("background-color", function(d) { return d; })
    .append("text")
    .text(function(d) {
      return color.invertExtent(d).join("-");
    })

    d3.json("us-states.json", function(json) {

      for (var i = 0; i < data.length; i++) {
        var dataState = data[i].state;
        for (var j = 0; j < json.features.length; j++) {
          var jsonState = json.features[j].properties.name;

          if (dataState == jsonState) {
            data[i].track_id = parseInt(data[i].track_id, 10)
            data[i].playcount = parseInt(data[i].playcount, 10)

            var props = json.features[j].properties;
            for (var k in data[i]) {
              props[k] = data[i][k];
            }

            break;
          }
        }
      }

      svg.selectAll("path")
      .data(json.features)
      .enter()
      .append("path")
      .attr("d", path)
      .style("fill", function(d) {
        var count = d.properties.count;
        if (count) {
          return color(count);
        } else {
          return "#ccc";
        }
      })
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

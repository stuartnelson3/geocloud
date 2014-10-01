window.onload = function() {
  var w = window.innerWidth;
  var h = window.innerHeight;
  function iframe(url) {
    return "<iframe width='100%' height='200' scrolling='no' frameborder='no' src='" + url + "'></iframe>"
  }
  function img(imgUrl, linkUrl) {
    return "<img class='album-art embed-link pointer' src='" + imgUrl + "' data-url='" + linkUrl + "' />";
  }

  $(document).on("click", ".embed-link", function() {
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
    if (!d.properties.tracks) {
      return "<div class='album-description'>No track data :(</div>";
    }
    return d.properties.tracks.slice(0,3).map(function(t) {
      if (t.artwork_url) {
        return "<div class='album-container'>" + img(t.artwork_url, t.permalink_url) +
          // "<div class='album-description'>" + t.title + "</div>" +
          // "<div class='album-description'>" + t.count + " plays</div>" +
          "</div>";
      } else {
        return "<div>No data found</div>";
      }
    }).join("");
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

  d3.json("/public/states.json", function(data) {
    data.forEach(function(state) {
      state.tracks.sort(function compare(a, b) {
        if (a.count < b.count) {
          return 1;
        }
        if (a.count > b.count) {
          return -1;
        }
        return 0;
      });
    });

    var colorScale = ["rgb(237,248,233)","rgb(186,228,179)","rgb(116,196,118)","rgb(49,163,84)","rgb(0,109,44)"];

    var color = d3.scale.quantize()
    .range(colorScale).domain([
      d3.min(data, function(d) { return d.total_plays; }),
      d3.max(data, function(d) { return d.total_plays; })
    ]);

    var format = d3.format("s");
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
      return color.invertExtent(d).map(function(n) { return format(Math.round(n*100)/100); }).join("-");
    })

    d3.json("/public/us-states.json", function(json) {

      for (var i = 0; i < data.length; i++) {
        var dataState = data[i].name;
        for (var j = 0; j < json.features.length; j++) {
          var jsonState = json.features[j].properties.name;

          if (dataState == jsonState) {
            var props = json.features[j].properties;
            for (var k in data[i]) {
              props[k] = data[i][k];
            }

            break;
          }
        }
      }

      var paths = svg.selectAll("path")
      .data(json.features);

      paths.enter()
      .append("path")
      .attr("d", path)
      .style("fill", setFill)
      .on("click", function(d) {
        appendStateContainer(d);
      })
      .on("mouseover", function(d) {
        d3.select(this).style("fill", "rgb(204, 68, 0)");
        tip.show(d)
      })
      .on("mouseout", function(d) {
        if ($(".d3-tip").is(":hover")) {
          return;
        } else {
        d3.select(this).style("fill", setFill);
          tip.hide(d)
          return;
        }
      });

      $(document).on("mouseleave", ".d3-tip", function() {
        tip.hide()
        paths.style("fill", setFill);
      });

      function setFill(d) {
        var total_plays = d.properties.total_plays;
        if (total_plays) {
          return color(total_plays);
        } else {
          return "#ccc";
        }
      }

      function appendStateContainer(d) {
        var s = d.properties;
        var el = document.querySelector(".state-data-container");
        var markup = ""
        markup = "<div class='state-name'>" + s.name + "</div>";
        markup += "<ol class='tracks'>";
        s.tracks.forEach(function(t) {
          markup += "<li class='track'>";
          markup += trackMarkup(t);
          markup += "</li>";
        });
        markup += "</ol>";
        el.innerHTML = markup;
      }

      function trackMarkup(t) {
        return "<span class='embed-link pointer' data-url='" + t.permalink_url + "'>" + t.title + "</span>";
      }

    });
  });
}

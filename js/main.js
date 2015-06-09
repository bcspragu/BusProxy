var ractive = new Ractive({
  // The `el` option can be a node, an ID, or a CSS selector.
  el: '.main-content',
  template: '#board',
  data: { 
    animation: "fade",
    stops: [],
    moment: moment,
    sProp: "",
    lProp: "",
    aProp: "",
  }
});

$(function() {
  updateBoard();
  console.log(setInterval(updateBoard, 10000));
});

function updateBoard() {
  $.get('/?stop=64&stop=63', function(stops) {
    var sProp = "col-xs-24 col-sm-2 col-md-2 col-lg-1";
    var lProp = "col-xs-24 col-sm-15 col-md-15 col-lg-16";
    var aProp = "col-xs-24 col-sm-7 col-md-7 col-lg-7";
    if (stops.length > 1) {
      sProp = "col-xs-24 col-sm-2 col-md-1 col-lg-2";
      lProp = "col-xs-24 col-sm-15 col-md-16 col-lg-15";
      aProp = "col-xs-24 col-sm-7 col-md-7 col-lg-7";
    }
    ractive.set('stops', stops);
    ractive.set('sProp', sProp);
    ractive.set('lProp', lProp);
    ractive.set('aProp', aProp);
  }, 'json');
}

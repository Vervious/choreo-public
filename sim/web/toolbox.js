
window.addEventListener("load", function () {
  function sendData() {
    var XHR = new XMLHttpRequest();

    // Bind the FormData object and the form element
    // console.log(form);
    var FD = new FormData(form);
    // console.log(FD);

for (var [key, value] of FD.entries()) { 
  console.log(key, value);
}

    // Define what happens on successful data submission
    XHR.addEventListener("load", function(event) {
      // alert(event.target.responseText);
      console.log("success");
      textbox.innerHTML = event.target.responseText;

      setTimeout(function() {
        textbox.innerHTML = "ready";
      }, 1000);
    });

    // Define what happens in case of error
    XHR.addEventListener("error", function(event) {
      alert("Failed to post toolbox. check console.");
    });

    // Set up our request
    XHR.open("POST", "http://localhost:8910/toolbox");

    // The data sent is what the user provided in the form
    XHR.send(FD);
  }
 
  // Access the form element...
  var form = document.getElementById("toolbox");
  var textbox = document.getElementById("toolbox-notes");
  console.log(textbox);

  // ...and take over its submit event.
  form.addEventListener("submit", function (event) {
    event.preventDefault();

    sendData();
  });
});
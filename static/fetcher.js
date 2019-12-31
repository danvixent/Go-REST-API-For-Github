function fetch() {
    var form = document.getElementById("frm")
    var divs = document.getElementById("response");

    form.addEventListener("submit", function (e) {
      e.preventDefault()

      var x = new XMLHttpRequest()

      x.onreadystatechange = function () {
        if (x.readyState == 4) {
          divs.innerHTML = x.responseText;
        }
      }

      x.open("POST", "/post")
      x.send(new FormData(form))
    })
  }
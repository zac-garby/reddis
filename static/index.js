function loadPosts(id) {
    var req = new XMLHttpRequest()

    req.onreadystatechange = function() {
        if (this.readyState == 4 && this.status == 200) {
            var elem = document.getElementById(`depth-${id}`)
            elem.outerHTML = this.responseText
        }
    }

    req.open("GET", `get_posts?id=${id}`, true)
    req.send()
}

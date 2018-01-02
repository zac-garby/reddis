function loadPosts(id) {
    var req = new XMLHttpRequest()

    req.onreadystatechange = function() {
        if (this.readyState == 4 && this.status == 200) {
            var elem = document.getElementById(`depth-${id}`)
            elem.outerHTML = this.responseText
        }
    }

    req.open('GET', `get_posts?id=${id}`, true)
    req.send()
}

function userExists(name) {
    return get(`user_exists?name=${name}`) === 'true'
}

function get(url) {
    var req = new XMLHttpRequest()
    var resp = 'no response'

    req.onreadystatechange = function() {
        if (this.readyState == 4 && this.status == 200) {
            resp = this.responseText
        }
    }

    req.open('GET', url, false)
    req.send()

    return resp
}

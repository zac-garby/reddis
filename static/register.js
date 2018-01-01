function beforeSubmit(form) {
    var name = form['name'].value
    var pass = form['password'].value
    var repeat = form['password-repeat'].value

    if (pass !== repeat) {
        document.getElementById('message').innerHTML = 'The two passwords don\'t match.'
        return false
    }

    if (pass.length === 0) {
        document.getElementById('message').innerHTML = 'Your password must be <strong>at least one character.</strong>'
        return false
    }

    if (!/^[a-zA-Z0-9-_]+$/.test(name)) {
        document.getElementById('message').innerHTML = `
            Invalid username. Can only contain
            <strong>alphanumeric characters, hyphens,
            and underscores</strong<, and must be <strong>at least
            one character.</strong>`
        return false
    }

    if (userExists(name)) {
        document.getElementById('message').innerHTML = `
            User <strong>${name}</strong> already exists. Note
            that usernames are case
            <strong>sensitive</strong>, so you
            can have the same name with different
            capitilisation.`
        return false
    }

    form['hash'].value = sha512(pass)

    return true
}

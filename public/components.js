'use strict'

class C_AuthBox {

    constructor (el) {
        this.el = el
        this.update = this.update.bind(this)
        this.submit = this.submit.bind(this)

        events.subscribe('ws_auth', 'authbox', this.update)
        this.create()
    }

    create () {
        this.selector = document.createElement('div')
        this.selector.className = 'selector'
        this.selector.innerHTML = ``
        this.form = document.createElement('form')
        this.form.onsubmit = this.submit
        this.form.innerHTML = `
            <label>Username</label> <input name="username" autocomplete="off">
            <label>Password</label> <input name="password" type="password">
            <input type="submit" value="Login">`
        this.el.appendChild(this.form)
        this.profile = document.createElement('div')
        this.profile.className = 'profile'
        this.profile.style.display = 'none'
        //this.profile.innerHTML = `<button>Logout</button>`
        this.el.appendChild(this.profile)
    }

    submit (evt) {
        fetch('//localhost:8080/api/auth', {
                method: 'POST',
                body: new FormData(this.form)
            }).then(r => {
                return r.json()
            }).then(json => {
                if (!json.error) {
                    connector.auth(json.data.token)
                } else {
                    throw new Error(json.message)
                }
            }).catch(err => {
                console.error('auth error', err)
            })
            return false
    }

    update (data) {
        if (!data.error && data.message == 'authenticated') {
            this.profile.innerHTML = `Hello ${data.data.name}!`
            this.profile.style.display = 'block'
            this.form.style.display = 'none'
        }
    }

}

class C_ConnectorStatus {

    constructor (el) {
        this.el = el
        this.update = this.update.bind(this)

        events.subscribe('ws_debug', 'panel', this.update)
        this.update()
    }

    update (data) {
        this.el.innerHTML = `
            <strong>State:</strong>${connector.ws ? connector.ws.readyState : 0}<br>
            <strong>Retries:</strong>${connector.retries}<br>
            <strong>Ready:</strong>${connector.ready}<br>
            <strong>Authenticated:</strong>${connector.authenticated}
        `
    }

}
'use strict'

const RETRY_WAIT = 1000
const MAX_RETRIES = 3
const ADDR = "ws://localhost:8080/api/ws"

class Connector {

    constructor () {
        this.retries = 0
        this.queue = []
        this.ready = false
        this.authenticated = false
    }

    open () {
        if (this.retries > MAX_RETRIES) throw new Error('Max retries hit')
        // if we're currently connecting or connecting, don't attempt
        if (this.ws != null && this.ws.readyState < 2) return
        this.ws = new WebSocket(ADDR)

        this.ws.onopen = (evt) => {
            console.log('ws opened', ADDR)
            this.retries = 0
            events.publish('ws_debug', null)
        }

        this.ws.onmessage = (evt) => {
            console.log('ws message', evt)
            this.receive(evt)
            events.publish('ws_debug', null)
        }

        this.ws.onerror = (evt) => {
            console.error('ws error', evt)
            this.retries++
            events.publish('ws_debug', null)
            this.open()
        }

        this.ws.onclose = (evt) => {
            console.log('ws closed')
            this.authenticated = false
            this.ready = false
            events.publish('ws_debug', null)
            this.open()
        }
    }

    close () {
        this.ws.close()
    }

    auth (token) {
        if (token) {
            this.token = token
        }
        if (this.token && this.ready && !this.authenticated) {
            this.ws.send(JSON.stringify({
                action: 'auth',
                data: {
                    token: this.token
                }
            }))
        }
    }

    flushQueue () {
        const queue = this.queue.slice()
        this.queue = []
        queue.forEach(d => this.send(d))
    }

    send (data) {
        if (!this.ws || this.ws.readyState != 1 ||
            !this.ready || !this.authenticated) {
            this.queue.push(data)
        } else {
            console.log('sending', data)
            this.ws.send(JSON.stringify(data))
        }
    }

    receive (evt) {
        let data
        try {
            data = JSON.parse(evt.data)
        } catch (err) {
            console.error('ignoring message, not JSON', evt)
            return
        }

        events.publish('ws_debug', data)

        if (data.error) {
            console.error('server error', data)
        }

        if (!this.ready && !this.authenticated) {
            if (!data.error && data.message == 'ok!') {
                this.ready = true
                this.auth()
            }
            events.publish('ws_status', this.ready)
        } else if (!this.authenticated) {
            if (!data.error && data.message == 'authenticated') {
                this.authenticated = true
            }
            events.publish('ws_auth', data)
        } else {
            events.publish('ws_message', data)
        }
    }

}
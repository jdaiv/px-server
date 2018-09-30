'use strict'

class StateManager {

}

class EventManager {

    constructor () {
        this.subscribers = {}
    }

    subscribe (type, id, func) {
        if (this.subscribers[type] == null) {
            this.subscribers[type] = {}
        }
        if (this.subscribers[type][id]) {
            console.warn(id, 'already subscibed to', type)
            return
        }
        this.subscribers[type][id] = func
    }

    unsubscribe (id) {
        if (this.subscribers[type] != null) {
            delete this.subscribers[type][id]
        }
    }

    publish (type, data) {
        const group = this.subscribers[type]
        if (group == null) {
            console.warn('no subscribers for (', type, '), payload ignored', data)
            return
        }
        for (let key in group) {
            if (group[key]) group[key](data)
        }
    }

}
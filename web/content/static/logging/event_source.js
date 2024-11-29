/** @type {EventSource|null} */
let eventSource = null

export function Open() {
  eventSource = new EventSource("/logs/events")
  eventSource.onopen = onOpen
  eventSource.onmessage= onMessage
  eventSource.onerror = onError
}

/**
 * @this {EventSource}
 * @param {Event} event
 */
function onOpen(event) {
  console.debug("opened event source", this, event)
}

/**
 * @this {EventSource}
 * @param {MessageEvent<any>} event
 */
function onMessage(event) {
  console.debug("received event source message", this, event)
}

/**
 * @this {EventSource}
 * @param {Event} event
 */
function onError(event) {
  if (this.readyState == this.CLOSED) {
    console.debug("event source closed", this, event)
    return 
  }

  console.debug("received event source error", this, event)
  alert("There was an error while receiving updates from the server. Check the console for details.")

  this.close()
}


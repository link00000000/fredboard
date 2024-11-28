export class LogEventSource {
  /** @type {EventSource|null} */
  #eventSource = null

  /**
   * @this {EventSource}
   * @param {Event} event
   */
  #onOpen(event) {
    console.debug("opened event source", this, event)
  }

  /**
   * @this {EventSource}
   * @param {MessageEvent<any>} event
   */
  #onMessage(event) {
    console.debug("recenved event source message", this, event)
  }

  /**
   * @this {EventSource}
   * @param {Event} event
   */
  #onError(event) {
    console.debug("received event source error", this, event)
    alert("There was an error while receiving updates from the server. Check the console for details.")
  }

  static Open() {
    const src = new LogEventSource()
    src.#eventSource = new EventSource("/events/logs")

    src.#eventSource.onopen = src.#onOpen
    src.#eventSource.onmessage = src.#onMessage
    src.#eventSource.onerror = src.#onError

    return src
  }

  Close() {
    this.#eventSource.onopen = null
    this.#eventSource.onmessage = null
    this.#eventSource.onerror = null

    this.#eventSource.close()
    this.#eventSource = null
  }
}


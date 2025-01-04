import { CSS } from "../lib/css.js"

/**
  * @typedef {"debug"|"info"|"warn"|"error"|"fatal"|"panic"} LogLevel
  */

/**
  * @param {string} str
  * @returns {LogLevel|null}
  */
function makeLogLevel(str) {
  switch (str) {
    case "debug":
    case "info":
    case "warn":
    case "error":
    case "fatal":
    case "panic":
      return str
    default:
      return null
  }
}

export class FredboardRecordElement extends HTMLElement {
  static tag = "fredboard-record"

  static attrs = {
    time: "time",
    level: "level",
    message: "message",
  }

  static observedAttributes = [FredboardRecordElement.attrs.time, FredboardRecordElement.attrs.level, FredboardRecordElement.attrs.message]

  static #css = CSS`
    .root {
      display: flex;
      gap: 8px;
    }

    .time {
      font-size: 0.6rem;
      color: gray;
    }

    .level {
      --level-color: #ffffff;
      
      font-size: 0.6rem;
      font-weight: bold;
      outline: 1px solid var(--level-color);
      color: var(--level-color);
      border-radius: 2px;
      padding: 2px;
    }

    .level[data-level="debug"] {
      --level-color: purple;
    }

    .level[data-level="info"] {
      --level-color: blue;
    }

    .level[data-level="warn"] {
      --level-color: yellow;
    }

    .level[data-level="error"] {
      --level-color: red;
    }

    .level[data-level="fatal"] {
      --level-color: red;

      color: white;
      background-color: var(--level-color);
    }

    .level[data-level="panic"] {
      --level-color: black;

      color: white;
      background-color: var(--level-color);
    }

    .message {

    }
  `

  #initialized = false

  /** @type {HTMLElement|undefined} */
  #rootElem

  /** @type {HTMLElement|undefined} */
  #timeElem

  /** @type {HTMLElement|undefined} */
  #levelElem

  /** @type {HTMLElement|undefined} */
  #messageElem

  // TODO: #contextElem

  time() {
    const attr = this.getAttribute(FredboardRecordElement.attrs.time)

    if (attr === null) {
      return null
    }

    const time = Date.parse(attr)
    if (isNaN(time)) {
      console.warn(`<fredboard-record /> has invalid "time" attribute "${attr}"`)
      return null
    }

    try {
      return new Date(time)
    } catch (error) {
      console.warn(`<fredboard-record /> has invalid "time" attribute "${attr}"`)
      return null
    }
  }

  /**
   * @returns {"debug"|"info"|"warn"|"error"|"fatal"|"panic"|null}
   */
  level() {
    const attr = this.getAttribute(FredboardRecordElement.attrs.level)

    if (attr === null) {
      return null
    }

    return makeLogLevel(attr)
  }

  message() {
    return this.getAttribute(FredboardRecordElement.attrs.message)
  }

  connectedCallback() {
    const shadow = this.attachShadow({ mode: "open" })

    const styleElem = document.createElement("style")
    styleElem.textContent = FredboardRecordElement.#css
    shadow.appendChild(styleElem)

    this.#rootElem = document.createElement("div")
    this.#rootElem.setAttribute("class", "root")
    shadow.appendChild(this.#rootElem)

    this.#timeElem = document.createElement("span")
    this.#timeElem.setAttribute("class", "time")
    this.#rootElem.appendChild(this.#timeElem)

    this.#levelElem = document.createElement("span")
    this.#levelElem.setAttribute("class", "level")
    this.#rootElem.appendChild(this.#levelElem)

    this.#messageElem = document.createElement("span")
    this.#messageElem.setAttribute("class", "message")
    this.#rootElem.appendChild(this.#messageElem)

    this.#initialized = true

    this.update()
  }

  /**
   * @param {string} _name
   * @param {string} _oldValue
   * @param {string} _newValue
   */
  attributeChangedCallback(_name, _oldValue, _newValue) {
    if (!this.#initialized) {
      return
    }

    this.update()
  }

  update() {
    if (this.#timeElem) {
      const time = this.time()
      this.#timeElem.textContent = time !== null ? time.toString() : ""
    }

    if (this.#levelElem) {
      const level = this.level()
      this.#levelElem.dataset.level = level !== null ? level : undefined
      this.#levelElem.textContent = level
    }

    if (this.#messageElem) {
      const message = this.message()
      this.#messageElem.textContent = message
    }
  }
}

customElements.define(FredboardRecordElement.tag, FredboardRecordElement)

export class FredboardLoggerElement extends HTMLElement {
  static tag = "fredboard-logger"

  static attrs = {
    loggerId: "logger-id",
  }

  /**
   * @param {string} loggerId
   * @returns {FredboardLoggerElement}
   */
  static createElement(loggerId) {
    /** @type FredboardLoggerElement */
    const elem = document.createElement(FredboardLoggerElement.tag)
    elem.setAttribute(FredboardLoggerElement.attrs.loggerId, loggerId)

    return elem
  }

  /**
   * @param {HTMLElement} scope
   * @param {string} id
   * @returns {FredboardLoggerElement|null}
   */
  static findByLoggerId(scope, id) {
    const selector = `${FredboardLoggerElement.tag}[${FredboardLoggerElement.attrs.loggerId}="${id}"]`
    const elem = scope.querySelector(selector)

    if (elem === null) {
      return null
    }

    if (!(elem instanceof FredboardLoggerElement)) {
      console.warn(`failed to find logger by id: element matching '${selector}' is not an instance of FredboardLoggerElement`)
      return null
    }

    return elem
  }
}

customElements.define(FredboardLoggerElement.tag, FredboardLoggerElement)

export class FredboardLogElement extends HTMLElement {
  static tag = "fredboard-log"

  /** @param {import("pages/logs.js").LoggerCreatedMessage} data */
  notifyLoggerCreated(data) {
    console.debug(`<${FredboardLogElement.tag}}> notified of logger created`, this, data)

    let existingElement = FredboardLoggerElement.findByLoggerId(this, data.logger.id)
    if (existingElement !== null) {
      console.warn(`tried to create <${FredboardLogElement.tag}> with loggerId="${data.logger.id}", but element already exists for that loggerId`)
      return
    }

    this.appendChild(FredboardLoggerElement.createElement(data.logger.id))
  }

  /** @param {import("pages/logs.js").LoggerClosedMessage} data */
  notifyLoggerClosed(data) {
    console.debug("<fredboard-log> notified of logger closed", this, data)
    // TODO
  }

  /** @param {import("pages/logs.js").RecordMessage} data */
  notifyRecord(data) {
    console.debug("<fredboard-log> notified of record", this, data)
    // TODO
  }
}

customElements.define(FredboardLogElement.tag, FredboardLogElement)

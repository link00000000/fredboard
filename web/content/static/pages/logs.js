"use strict";

import * as errors from "../lib/errors.js"

const MessageType = {
  LoggerCreated: 0,
  LoggerClosed: 1,
  Record: 2,
}

const eventSource = new EventSource("/logs/events")
eventSource.onopen = onOpen
eventSource.onmessage = onMessage
eventSource.onerror = onError

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

  try {
    const payload = JSON.parse(event.data)

    switch (payload.type) {
    case MessageType.LoggerCreated: {
      const [data, ok] = parseLoggerCreatedMessage(payload.data)

      if (!ok) {
        console.warn("ignoring malformed message", this, event, payload)
        break
      }

      onLoggerCreated(data)
      break
    }
    case MessageType.LoggerClosed: {
      const [data, ok] = parseLoggerClosedMessage(payload.data)

      if (!ok) {
        console.warn("ignoring malformed message", this, event, payload)
        break
      }

      onLoggerClosed(payload.data)
      break
    }
    case MessageType.Record: {
      const [data, error] = parseRecordMessage(payload.data)

      if (error !== null) {
        console.warn("ignoring malformed message", this, event, payload, error)
        break
      }

      onRecord(data)
      break
    }
    default:
      console.warn("ignoring malformed message", this, event, payload)
    }
  }
  catch (error) {
    console.error("error while handling message", this, event, error)
  }
}

/**
 * @this {EventSource}
 * @param {Event} event
 */
function onError(event) {
  if (this.readyState === this.CLOSED) {
    console.debug("event source closed", this, event)
  }
  else {
    console.error("received event source error", this, event)
    alert("There was an error while receiving updates from the server. Check the console for details.")
  }

  this.close()
}

/**
 * @typedef {object} Caller
 * @property {string} file
 * @property {number} line
 */

/**
 * @typedef {object} Logger
 * @property {string} id
 * @property {string|undefined} parent
 * @property {string[]} children
 * @property {string} root
 */

/**
 * @typedef {object} LoggerCreatedMessage
 * @property {Date} time
 * @property {Caller} caller
 * @property {Logger} logger
 */

/**
 * @typedef {object} LoggerClosedMessage
 * @property {Date} time
 * @property {Caller} caller
 * @property {Logger} logger
 */

/**
 * @typedef {object} RecordMessage
 * @property {Date} time
 * @property {"debug"|"info"|"warn"|"error"|"fatal"|"panic"} level
 * @property {string} message
 * @property {string|undefined} error
 * @property {Caller|undefined} caller
 * @property {Logger} logger
 */

/**
 * @param {unknown} data
 * @returns {[LoggerCreatedMessage, boolean]}
 */
function parseLoggerCreatedMessage(data) {
  // TODO
}

/**
 * @param {unknown} data
 * @returns {[LoggerClosedMessage, boolean]}
 */
function parseLoggerClosedMessage(data) {
  // TODO
}

/**
 * @param {unknown} data
 * @returns {[RecordMessage, null]|[null, Error]}
 */
function parseRecordMessage(data) {
  /** @type RecordMessage */
  let msg = {}

  if (typeof data !== "object") {
    return [null, new Error("'data' is not of type 'object'")]
  }

  if (data === null) {
    return [null, new Error("'data' is null")]
  }

  if (!("time" in data && typeof data.time === "string")) {
    return [null, new Error("property 'time' not of type 'string'")]
  }

  const parsedTime = Date.parse(data.time)
  if (isNaN(parsedTime)) {
    return [null, new Error("property 'time' could not be parsed as a Date object")]
  }

  const [time, error] = errors.trycatch(() => new Date(parsedTime))
  if (error != null) {
    return [null, new Error("property 'time' could not be parsed as a Date object: " + error.message)]
  }

  msg.time = time

  if (!("level" in data && typeof data.level === "string")) {
    return [null, new Error("property 'level' not of type 'string'")]
  }

  switch (data.level) {
    case "debug":
    case "info":
    case "warn":
    case "error":
    case "fatal":
    case "panic":
      break
    default:
      return [null, new Error("property 'level' has invalid value " + data.level.toLowerCase())]
  }

  msg.level = data.level

  if (!("message" in data && typeof data.message === "string")) {
    return [null, new Error("property 'message' not of type 'string'")]
  }

  msg.message = data.message

  if ("error" in data && data.error !== null) {
    if (!(typeof data.error === "string")) {
      return [null, new Error("property 'error' not of type 'string' or null")]
    }

    msg.error = data.error
  }

  if ("caller" in data && data.caller !== null) {
    if (!(typeof data.caller === "object")) {
      return [null, new Error("property 'caller' not of type 'object' or null")]
    }

    if (!("file" in data.caller && typeof data.caller.file === "string")) {
      return [null, new Error("property 'caller.file' not of type 'string'")]
    }

    if (!("line" in data.caller && typeof data.caller.line === "number")) {
      return [null, new Error("property 'caller.line' not of type 'number'")]
    }

    if (!(Number.isInteger(data.caller.line) && data.caller.line > 0)) {
      return [null, new Error("property 'caller.line' is not a positive integer")]
    }

    msg.caller = { file: data.caller.file, line: data.caller.line }
  }

  if (!("logger" in data && data.logger !== null && typeof data.logger === "object")) {
    return [null, new Error("property 'logger' not of type 'object'")]
  }


  if (!("id" in data.logger && typeof data.logger.id === "string")) {
    return [null, new Error("property 'data.logger.id' not of type 'string'")]
  }

  /** @type {string|undefined} */
  let parentLogger = undefined

  if ("parent" in data.logger) {
    if (typeof data.logger.parent !== "string") {
      return [null, new Error("property 'data.logger.id' not of type 'string'")]
    }

    parentLogger = data.logger.parent
  }

  if (!("children" in data.logger && typeof data.logger.children === "object" && Array.isArray(data.logger.children))) {
    return [null, new Error("property 'data.logger.children' not of type 'array'")]
  }


  /** @type {string[]} */
  let childLoggers = []

  for (let i = 0; i < data.logger.children.length; ++i) {
    const child = data.logger.children[i]
    if (typeof child !== "string") {
      return [null, new Error(`property 'data.logger.children[${i}]' not of type 'string'`)]
    }

    childLoggers.push(child)
  }

  if (!("root" in data.logger && typeof data.logger.root === "string")) {
    return [null, new Error("property 'data.logger.root' not of type 'string'")]
  }

  msg.logger = { id: data.logger.id, parent: parentLogger, children: childLoggers, root: data.logger.root }

  return [msg, null]
}

/**
 * @param {LoggerCreatedMessage} data
 */
function onLoggerCreated(data) {
  // TODO
}

/**
 * @param {LoggerClosedMessage} data
 */
function onLoggerClosed(data) {
  // TODO
}

/**
 * @param {RecordMessage} data
 */
function onRecord(data) {
  console.info("new record", data)
}


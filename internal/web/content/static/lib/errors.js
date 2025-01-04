/**
  * @param {unknown} caught
  * @returns {Error}
  */
export function makeError(caught) {
  if (caught instanceof Error) {
    return caught
  }

  if (caught instanceof Object) {
    return new Error(Object.toString())
  }

  return new Error("unknown error")
}

/**
  * @template T
  * @param {() => T} predicate
  * @returns {[T, null]|[null, Error]}
  */
export function trycatch(predicate) {
  try {
    return [predicate(), null]
  }
  catch (error) {
    return [null, makeError(error)]
  }
}

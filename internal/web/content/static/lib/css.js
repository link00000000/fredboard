/**
  * Exists to prepend CSS to a string literal to hint to
  * the editor to highlight the string as CSS
  * @param {TemplateStringsArray} strings
  */
export function CSS(strings) {
  return strings.join("\n")
}

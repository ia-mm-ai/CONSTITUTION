const ESCAPES = Object.freeze({
  "<": "\\u003c",
  ">": "\\u003e",
  "&": "\\u0026",
  "\u2028": "\\u2028",
  "\u2029": "\\u2029"
});

// Go json.Marshal escapes these characters in signed and submitted envelopes.
export function goJSONStringify(value) {
  return JSON.stringify(value).replace(/[<>&\u2028\u2029]/g, (character) => ESCAPES[character]);
}

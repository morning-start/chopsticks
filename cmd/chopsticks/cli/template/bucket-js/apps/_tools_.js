/**
 * @global
 * @param {string} s
 * @returns {string}
 */
let trim = (s) => s.trim();

/**
 * @global
 * @param {string} s
 * @param {string} delimiter
 * @returns {string[]}
 */
let split = (s, delimiter) => {
    return s.split(delimiter);
};

/**
 * @global
 * @param {string} str
 * @param {string} prefix
 * @returns {boolean}
 */
let startsWith = (str, prefix) => {
    return str.startsWith(prefix);
};

/**
 * @global
 * @param {string} str
 * @param {string} suffix
 * @returns {boolean}
 */
let endsWith = (str, suffix) => {
    return str.endsWith(suffix);
};

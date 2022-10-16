const pipe =
  (...fns) =>
  (x) =>
    fns.reduce((v, f) => f(v), x);

const curry = (func) => {
  const curried = (...args) => {
    if (args.length >= func.length) {
      return func.apply(this, args);
    }
    return (...args2) => {
      return curried.apply(this, args.concat(args2));
    };
  };
  return curried;
};

// util for point free variants (composable)
const splitOnComma = (str) => str.split(",");
const toString = (int) => int.toString();

// String Transformers. Expects `string` returns `string`
const toLowerCase = (str) => str.toLowerCase();
const trim = (str) => str.trim();
const normaliseUTF8 = (str) => str.normalize("NFD"); // standises UTF8 character codes
const normaliseDiacritics = (str) => str.replace(/[\u0300-\u036f]/g, ""); // normalise diacritics "éàçèñ" -> "eacen"
const normaliseString = pipe(
  trim,
  toLowerCase,
  normaliseUTF8,
  normaliseDiacritics
);

// Array Utils
const normaliseArray = (a) => a.map(normaliseString); // standises UTF8 codes
const onDuplicate = (s, i, a) => a.indexOf(s) === i;
const dedupe = (a) => a.filter(onDuplicate);

// Object mappers
const normaliseTag = (tag) => normaliseString(tag.value);
const normaliseFlag = (flag) => ({
  id: toString(flag.id),
  description: normaliseString(flag.description),
  tags: flag.tags.map(normaliseTag).join(" "),
});

// Flag matchers
const propertyValueMatcher = curry((property, value, flag) => {
  if (!flag[property]) {
    return false;
  }
  return flag[property].includes(value);
});

const defaultMatcher = curry((term, flag) => {
  return (
    flag.id.includes(term) ||
    flag.description.includes(term) ||
    flag.tags.includes(term)
  );
});

const applyMatchers = curry((matcherFns, flag) =>
  matcherFns.map((matcher) => matcher(flag)).includes(true)
);

// Tokeniser, basic comma seperated lists. Expects `string` returns `string[]`
const tokenise = pipe(splitOnComma, normaliseArray, dedupe);

const filterFlags = (haystack, input) => {
  const terms = tokenise(input);
  // build terms into array of patially applied matchers
  const matcherFns = terms.map((term) => {
    if (term.includes(":")) {
      const [property, value] = term.split(":");
      return propertyValueMatcher(property, value);
    }

    return defaultMatcher(term);
  });

  const results = haystack
    .map(normaliseFlag)
    .filter(applyMatchers(matcherFns))
    .map((flag) => parseInt(flag.id, 10)); // collect matched IDs
  return haystack.filter((flag) => results.includes(flag.id));
};

export default {
  filterFlags,
};

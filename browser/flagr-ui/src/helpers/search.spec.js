// import { filterFlags } from "./search";
import search from "./search";
const { filterFlags } = search;

const mockFlags = [
  {
    id: 1,
    tags: [{ value: "foo" }],
    description: "bar",
  },
  {
    id: 2,
    tags: [{ value: "sun" }],
    description: "moon luna",
  },
  {
    id: 3,
    tags: [{ value: "sun" }, { value: "solar" }],
    description: "bar-ight 2",
  },
];

describe("filterFlags", () => {
  it("should filter given a string", () => {
    const result = filterFlags(mockFlags, "foo");

    expect(result.length).toEqual(1);
    expect(result[0].id).toEqual(1);
  });
  it("should ignore case", () => {
    const result = filterFlags(mockFlags, "FOO");

    expect(result.length).toEqual(1);
    expect(result[0].id).toEqual(1);
  });
  it("should match numeric values", () => {
    const result = filterFlags(mockFlags, "1");

    expect(result.length).toEqual(1);
    expect(result[0].id).toEqual(1);
  });
  it("should partially match", () => {
    const result = filterFlags(mockFlags, "bar");

    expect(result.length).toEqual(2);
    expect(result[0].id).toEqual(1);
    expect(result[1].id).toEqual(3);
  });
  it("should match property and value given property:value", () => {
    const result = filterFlags(mockFlags, "id:2");

    expect(result.length).toEqual(1);
    expect(result[0].id).toEqual(2);
  });
  it("should match tags and value given property:value", () => {
    const result = filterFlags(mockFlags, "tags:solar");

    expect(result.length).toEqual(1);
    expect(result[0].id).toEqual(3);
  });
  it("should match mulitple terms", () => {
    const result = filterFlags(mockFlags, "id:1, id:3");

    expect(result.length).toEqual(2);
    expect(result[0].id).toEqual(1);
    expect(result[1].id).toEqual(3);
  });
  it("should ignore whitespace", () => {
    const result = filterFlags(mockFlags, "  luna  ");

    expect(result.length).toEqual(1);
    expect(result[0].id).toEqual(2);
  });
  it("should have no results if no match found", () => {
    const result = filterFlags(mockFlags, "garbage input");

    expect(result.length).toEqual(0);
  });
  it("should ignore input if not terms entered", () => {
    const result = filterFlags(mockFlags, "    ");

    expect(result.length).toEqual(3);
  });
  it("should match diacritics", () => {
    const result = filterFlags(mockFlags, "SÜN");

    expect(result.length).toEqual(2);
    expect(result[0].id).toEqual(2);
    expect(result[1].id).toEqual(3);
  });
});

export class Operators {
  static IN_OPERATOR = "IN";
  static NOTIN_OPERATOR = "NOTIN";
  static CHECKLIST_OPERATOR = "CHECKLIST";

  static ITEMS = [
    { value: "EQ", label: "==" },
    { value: "NEQ", label: "!=" },
    { value: "LT", label: "<" },
    { value: "LTE", label: "<=" },
    { value: "GT", label: ">" },
    { value: "GTE", label: ">=" },
    { value: "EREG", label: "=~" },
    { value: "NEREG", label: "!~" },
    { value: Operators.IN_OPERATOR, label: "IN" },
    { value: Operators.NOTIN_OPERATOR, label: "NOT IN" },
    { value: "CONTAINS", label: "CONTAINS" },
    { value: "NOTCONTAINS", label: "NOT CONTAINS" },
    { value: Operators.CHECKLIST_OPERATOR, label: "CHECK-LIST" },
  ].map(Object.freeze);

  static COLLECTION_VALUE_OPERATORS = Object.freeze([
    Operators.IN_OPERATOR,
    Operators.NOTIN_OPERATOR,
    Operators.CHECKLIST_OPERATOR,
  ]);

  static isCollectionValue(operator) {
    return Operators.COLLECTION_VALUE_OPERATORS.includes(operator);
  }

  set(operator) {
    this.operator = operator;
    this.setRule();
  }

  getLabel() {
    return Operators.ITEMS.find((item) => item.value === this.operator)?.label;
  }

  isCheckList() {
    return this.operator === Operators.CHECKLIST_OPERATOR;
  }

  setRule() {
    switch (this.operator) {
      case Operators.IN_OPERATOR:
      case Operators.NOTIN_OPERATOR:
        this.rule = arrayRule;
        break;
      case Operators.CHECKLIST_OPERATOR:
        this.rule = checklistRule;
        break;
      default:
        this.rule = null;
        break;
    }
  }

  new(lastItem) {
    return this.rule.new(lastItem);
  }

  unpack(value) {
    return JSON.parse(value || "[]").map(this.rule.unpack);
  }

  pack(value) {
    return JSON.stringify(value.map(this.rule.pack));
  }

  validateCollection(collection) {
    return this.rule?.validateCollection(collection);
  }

  validateCollectionNewItem(collection, item) {
    const foundIndex = collection.findIndex(collectionItem => collectionItem.value === item.value);
    if (~foundIndex) {
      return[`Value already exists in #${foundIndex + 1} row`];
    }

    return [];
  }
}

const arrayRule = Object.freeze({
  new: function () {
    return {
      value: "",
    };
  },
  pack: function (item) {
    return item.value;
  },
  unpack: function (item) {
    return {
      value: item,
    };
  },
  validateCollection: function (collection) {
    const result = { errors: [], details: {} };

    if (!collection.length) {
      result.errors.push("Collection cannot be empty");
      return result;
    }

    const detailsFlags = {
      empty: false,
      dups: false,
    };
    collection.forEach((item, index) => {
      const errors = [];
      
      if (!item.value) {
        errors.push('Value cannot be empty');
        detailsFlags.empty = true;
      }

      const foundIndex = collection.findIndex(collectionItem => collectionItem.value === item.value);
      if (foundIndex !== index) {
        errors.push(`Value already exists in #${foundIndex + 1} row`);
        detailsFlags.dups = true;
      }

      if (errors.length) {
        result.details[index] = errors;
      }
    });

    if (detailsFlags.empty) {
      result.errors.push("Collection contains empty value(s)");
    }

    if (detailsFlags.dups) {
      result.errors.push("Collection contains duplicated values");
    }

    return result;
  },
});

const checklistRule = Object.freeze({
  new: function (lastItem) {
    return {
      checked: lastItem?.checked !== false,
      value: "",
    };
  },
  pack: function (item) {
    return {
      c: +item.checked,
      v: item.value,
      d: !item.description ? 0 : item.description,
    };
  },
  unpack: function (item) {
    return {
      checked: item.c === 1,
      value: item.v,
      description: item.d === 0 ? "" : item.d,
    };
  },
  validateCollection: function (collection) {
    const result = arrayRule.validateCollection(collection);
    
    if (collection.length && collection.every(item => !item.checked)) {
      result.errors.push('Collection must contain at least one checked element');
    }

    return result;
  },
});

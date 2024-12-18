export const plural = (str) => {
    return str + "s";
}

export const reduceOperatorsToLabel = (operators) => {
    return operators.reduce((acc, el) => {
        acc[el.value] = el.label;
        return acc;
    }, {})
}
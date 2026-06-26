/** Bind flag-page handlers that take `(vm, …args)` to a Vue instance for template use. */
export function bindPageHandlers<Vm, T extends Record<string, (vm: Vm, ...args: never[]) => unknown>>(
  vm: Vm,
  handlers: T,
): { [K in keyof T]: T[K] extends (vm: Vm, ...args: infer A) => infer R ? (...args: A) => R : never } {
  const out = {} as { [K in keyof T]: T[K] extends (vm: Vm, ...args: infer A) => infer R ? (...args: A) => R : never }
  for (const key of Object.keys(handlers) as (keyof T)[]) {
    const fn = handlers[key]
    out[key] = ((...args: never[]) => fn(vm, ...args)) as (typeof out)[typeof key]
  }
  return out
}

/** Stub methods for `vue-tsc` (real impl bound in `created` via `bindPageHandlers`). */
export function pageMethodStubs<
  Vm,
  T extends Record<string, (vm: Vm, ...args: never[]) => unknown>,
>(
  _handlers: T,
): {
  [K in keyof T]: T[K] extends (vm: Vm, ...args: infer A) => infer R ? (...args: A) => R : never
} {
  const out = {} as {
    [K in keyof T]: T[K] extends (vm: Vm, ...args: infer A) => infer R ? (...args: A) => R : never
  }
  for (const key of Object.keys(_handlers) as (keyof T)[]) {
    out[key] = ((_args?: never) => undefined) as (typeof out)[typeof key]
  }
  return out
}
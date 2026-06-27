export function pluck<T, K extends keyof T>(arr: T[], prop: K): T[K][] {
  return arr.map((el) => el[prop])
}

export function sum(arr: number[]): number {
  return arr.reduce((acc, el) => acc + el, 0)
}

export function debounce<T extends (...args: never[]) => void>(
  fn: T,
  delay: number,
): (...args: Parameters<T>) => void {
  let timer: ReturnType<typeof setTimeout> | undefined
  return (...args: Parameters<T>) => {
    clearTimeout(timer)
    timer = setTimeout(() => fn(...args), delay)
  }
}

export default { pluck, sum, debounce }
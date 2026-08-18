// 每循环一次颜色降低百分之十
export function getColorArray(
  length: number = 0,
  baseColors: string[] = BASE_COLOR_ARRAY,
  initialOpacity: number = 1,
  opacityStep: number = 0.1,
): string[] {
  if (length <= baseColors.length) {
    return baseColors
  }

  const colorArray: string[] = []
  const baseColorCount = baseColors.length

  const totalCycles = Math.ceil(length / baseColorCount)

  for (let cycle = 1; cycle < totalCycles; cycle++) {
    const currentOpacity = Math.max(initialOpacity - cycle * opacityStep, 0.1)

    for (let index = 0; index < baseColorCount; index++) {
      if (colorArray.length >= length) {
        return colorArray
      }

      const baseColor = baseColors[index]
      const colorWithOpacity = addOpacityToHex(baseColor, currentOpacity)
      colorArray.push(colorWithOpacity)
    }
  }

  return colorArray
}

function addOpacityToHex(hex: string, opacity: number): string {
  const cleanHex = hex.replace('#', '')

  const fullHex =
    cleanHex.length === 3
      ? cleanHex
          .split('')
          .map((c) => c + c)
          .join('')
      : cleanHex

  const r = parseInt(fullHex.substring(0, 2), 16)
  const g = parseInt(fullHex.substring(2, 4), 16)
  const b = parseInt(fullHex.substring(4, 6), 16)

  return `rgba(${r}, ${g}, ${b}, ${opacity.toFixed(2)})`
}

export const BASE_COLOR_ARRAY = [
  '#07404B',
  '#75B83C',
  '#BCDC4A',
  '#B0D7C5',
  '#C6E5DE',
  '#B8BCBC',
]

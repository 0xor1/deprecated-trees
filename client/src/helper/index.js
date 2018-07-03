export function printDuration (duration, useAbs, hoursPerDay, daysPerWeek) {
  let minsPerYear = 525600
  let minsPerWeek = 10080
  let minsPerDay = 1440
  if (!useAbs) {
    minsPerDay = Math.floor(hoursPerDay * 60)
    minsPerWeek = Math.floor(minsPerDay * daysPerWeek)
    minsPerYear = Math.floor(minsPerWeek * 52)
  }
  if (duration === null || duration === undefined) {
    return '--'
  }
  let years = Math.floor(duration / minsPerYear)
  duration -= (years * minsPerYear)
  let weeks = Math.floor(duration / minsPerWeek)
  duration -= (weeks * minsPerWeek)
  let days = Math.floor(duration / minsPerDay)
  duration -= (days * minsPerDay)
  let hours = Math.floor(duration / 60)
  let minutes = duration - (hours * 60)
  if (minutes < 0) {
    minutes = 0
  }
  let res = ''
  if (years > 0) {
    res += years + 'y '
  }
  if (weeks > 0) {
    res += weeks + 'w '
  }
  if (days > 0) {
    res += days + 'd '
  }
  if (hours > 0) {
    res += hours + 'h '
  }
  if (res === '' || minutes > 0) {
    res += minutes + 'm'
  }
  return res
}

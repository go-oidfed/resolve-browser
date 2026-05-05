export function formatTime(unixTimestamp) {
  if (!unixTimestamp) return ''
  return new Date(unixTimestamp * 1000).toLocaleString()
}

export function formatEntityId(url) {
  try {
    const urlObj = new URL(url)
    return urlObj.hostname
  } catch {
    return url
  }
}

export function copyToClipboard(text) {
  navigator.clipboard.writeText(text)
}

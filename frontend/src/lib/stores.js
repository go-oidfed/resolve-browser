import { writable } from 'svelte/store'

export const trustAnchor = writable('')
export const subject = writable('')
export const entityTypes = writable([])
export const loading = writable(false)
export const error = writable(null)
export const result = writable(null)
export const originalResult = writable(null)
export const editedChain = writable([])
export const showDiff = writable(false)
export const previewMetadata = writable(null)

export function updateURL(trustAnchorValue, subjectValue, entityTypesValue) {
  const url = new URL(window.location.href)
  const params = url.searchParams
  
  if (trustAnchorValue) {
    params.set('ta', trustAnchorValue)
  } else {
    params.delete('ta')
  }
  
  if (subjectValue) {
    params.set('sub', subjectValue)
  } else {
    params.delete('sub')
  }
  
  if (entityTypesValue && entityTypesValue.length > 0) {
    params.set('types', entityTypesValue.join(','))
  } else {
    params.delete('types')
  }
  
  window.history.pushState({}, '', url.toString())
}

export function loadFromURL() {
  const params = new URLSearchParams(window.location.search)
  return {
    trustAnchor: params.get('ta') || '',
    subject: params.get('sub') || '',
    entityTypes: params.get('types') ? params.get('types').split(',') : []
  }
}

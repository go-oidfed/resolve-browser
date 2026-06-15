import Prism from 'prismjs'
import 'prismjs/components/prism-json'

export function highlightJson(data) {
  if (!data) return ''
  
  let jsonString
  try {
    jsonString = typeof data === 'string' ? data : JSON.stringify(data, null, 2)
  } catch (err) {
    return String(data)
  }
  
  try {
    const highlighted = Prism.highlight(jsonString, Prism.languages.json, 'json')
    return highlighted
  } catch (err) {
    return escapeHtml(jsonString)
  }
}

export function validateJson(str) {
  if (!str || str.trim() === '') {
    return { valid: true, error: null, line: null, column: null }
  }
  
  try {
    JSON.parse(str)
    return { valid: true, error: null, line: null, column: null }
  } catch (err) {
    const match = err.message.match(/position (\d+)/)
    if (match) {
      const position = parseInt(match[1], 10)
      const lines = str.substring(0, position).split('\n')
      const line = lines.length
      const column = lines[lines.length - 1].length + 1
      return { 
        valid: false, 
        error: err.message,
        line,
        column
      }
    }
    return { valid: false, error: err.message, line: null, column: null }
  }
}

export function formatJson(str) {
  if (!str || str.trim() === '') return ''
  
  try {
    const parsed = JSON.parse(str)
    return JSON.stringify(parsed, null, 2)
  } catch (err) {
    return str
  }
}

export function getThemeColors(darkMode) {
  if (darkMode) {
    return {
      background: '#1e293b',
      foreground: '#f1f5f9',
      gutterBackground: '#0f172a',
      gutterForeground: '#64748b',
      selection: '#334155',
      lineHighlight: '#1e293b',
      keyword: '#c678dd',
      string: '#98c379',
      number: '#d19a66',
      boolean: '#e06c75',
      null: '#abb2bf',
      key: '#61afef',
      brace: '#dcdfe4',
      error: '#e06c75',
      errorBackground: '#e06c7520'
    }
  }
  
  return {
    background: '#ffffff',
    foreground: '#1e293b',
    gutterBackground: '#f8fafc',
    gutterForeground: '#94a3b8',
    selection: '#e2e8f0',
    lineHighlight: '#f1f5f9',
    keyword: '#a626a4',
    string: '#059940',
    number: '#9c27b0',
    boolean: '#d32f2f',
    null: '#757575',
    key: '#0066cc',
    brace: '#424242',
    error: '#d32f2f',
    errorBackground: '#ffebee'
  }
}

function escapeHtml(str) {
  const div = document.createElement('div')
  div.textContent = str
  return div.innerHTML
}

export function parseJsonWithError(str) {
  if (!str || str.trim() === '') {
    return { success: true, data: null }
  }
  
  try {
    const data = JSON.parse(str)
    return { success: true, data }
  } catch (err) {
    return { success: false, error: err.message }
  }
}

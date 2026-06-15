<script>
  import { onMount, onDestroy } from 'svelte'
  import { EditorView, keymap, lineNumbers, highlightActiveLine, highlightActiveLineGutter } from '@codemirror/view'
  import { EditorState, Compartment } from '@codemirror/state'
  import { json } from '@codemirror/lang-json'
  import { oneDark } from '@codemirror/theme-one-dark'
  import { linter, lintKeymap } from '@codemirror/lint'
  import { defaultKeymap, history, historyKeymap, indentWithTab } from '@codemirror/commands'
  import { bracketMatching, foldGutter, foldKeymap } from '@codemirror/language'
  import { autocompletion } from '@codemirror/autocomplete'
  import { diagnosticCount, forceLinting } from '@codemirror/lint'
  import { getThemeColors, validateJson, formatJson } from '../lib/json-utils.js'
  
  export let value = ''
  export let onChange = (value) => {}
  export let readOnly = false
  export let placeholder = ''
  export let height = 'auto'
  
  let container
  let editorView = null
  let darkMode = false
  let hasErrors = false
  
  const themeCompartment = new Compartment()
  const readOnlyCompartment = new Compartment()
  let themeExtension = null
  let readOnlyExtension = null
  
  function checkDarkMode() {
    return document.documentElement.classList.contains('dark')
  }
  
  function getLinter() {
    return linter((view) => {
      const doc = view.state.doc.toString()
      const validation = validateJson(doc)
      
      if (!validation.valid && validation.error) {
        const pos = validation.line !== null 
          ? view.state.doc.line(validation.line).from + (validation.column || 1) - 1
          : 0
        
        return [{
          from: Math.max(0, pos),
          to: Math.min(view.state.doc.length, pos + 1),
          severity: 'error',
          message: validation.error
        }]
      }
      
      return []
    }, {
      delay: 300
    })
  }
  
  function createThemeExtension() {
    darkMode = checkDarkMode()
    const colors = getThemeColors(darkMode)
    
    const baseTheme = EditorView.theme({
      '&': {
        backgroundColor: colors.background,
        color: colors.foreground,
        fontSize: '14px',
        fontFamily: 'ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, "Liberation Mono", monospace'
      },
      '.cm-content': {
        padding: '8px 0',
        caretColor: colors.foreground
      },
      '.cm-line': {
        padding: '2px 8px',
        lineHeight: '1.5'
      },
      '.cm-gutters': {
        backgroundColor: colors.gutterBackground,
        color: colors.gutterForeground,
        borderRight: `1px solid ${darkMode ? '#334155' : '#e2e8f0'}`,
        paddingRight: '4px'
      },
      '.cm-activeLine': {
        backgroundColor: colors.lineHighlight
      },
      '.cm-activeLineGutter': {
        backgroundColor: darkMode ? '#1e293b' : '#f1f5f9'
      },
      '.cm-selectionBackground': {
        backgroundColor: colors.selection
      },
      '.cm-focused': {
        outline: 'none'
      },
      '.cm-scroller': {
        overflow: 'auto',
        maxHeight: height === 'auto' ? 'none' : height
      },
      '&.cm-editor.cm-focused': {
        outline: 'none'
      },
      '.cm-tooltip': {
        backgroundColor: darkMode ? '#1e293b' : '#ffffff',
        border: `1px solid ${darkMode ? '#334155' : '#e2e8f0'}`,
        borderRadius: '4px',
        boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)'
      },
      '.cm-tooltip-lint': {
        padding: '8px',
        fontSize: '12px'
      },
      '.cm-diagnostic': {
        padding: '4px 8px',
        margin: '2px 0',
        borderRadius: '4px',
        backgroundColor: colors.errorBackground,
        borderLeft: `3px solid ${colors.error}`
      },
      '.cm-diagnostic-error': {
        borderLeftColor: colors.error
      },
      '.cm-foldGutter': {
        width: '16px'
      },
      '.cm-foldGutter .cm-gutterElement': {
        color: colors.gutterForeground,
        cursor: 'pointer'
      },
      '.cm-placeholder': {
        color: darkMode ? '#64748b' : '#94a3b8'
      }
    }, { dark: darkMode })
    
    const customSyntaxHighlighting = EditorView.theme({
      '.cm-string': { color: colors.string },
      '.cm-number': { color: colors.number },
      '.cm-atom': { color: colors.boolean },
      '.cm-keyword': { color: colors.keyword },
      '.cm-property': { color: colors.key },
      '.cm-null': { color: colors.null },
      '.cm-bracket': { color: colors.brace }
    })
    
    return [baseTheme, customSyntaxHighlighting, oneDark]
  }
  
  function getExtensions() {
    themeExtension = createThemeExtension()
    readOnlyExtension = readOnly ? [EditorView.editable.of(false)] : []
    
    return [
      json(),
      getLinter(),
      lineNumbers(),
      foldGutter({
        openText: '▼',
        closedText: '▶'
      }),
      bracketMatching(),
      autocompletion({ override: [] }),
      highlightActiveLineGutter(),
      highlightActiveLine(),
      history(),
      keymap.of([
        ...defaultKeymap,
        ...historyKeymap,
        ...foldKeymap,
        ...lintKeymap,
        indentWithTab
      ]),
      themeCompartment.of(themeExtension),
      readOnlyCompartment.of(readOnlyExtension),
      EditorView.updateListener.of((update) => {
        if (update.docChanged) {
          const newValue = update.state.doc.toString()
          onChange(newValue)
          
          if (container && height === 'auto') {
            const contentHeight = container.querySelector('.cm-content')?.scrollHeight || 0
            const lineHeight = 21
            const lines = update.state.doc.lines
            const newHeight = Math.min(Math.max(contentHeight, lineHeight * 3), lineHeight * 20)
            container.style.height = `${newHeight}px`
          }
        }
        
        if (update.viewportChanged || update.docChanged) {
          hasErrors = diagnosticCount(update.view) > 0
        }
      }),
      EditorView.contentAttributes.of(placeholder ? { 'data-placeholder': placeholder } : {})
    ]
  }
  
  onMount(() => {
    const state = EditorState.create({
      doc: value,
      extensions: getExtensions()
    })
    
    editorView = new EditorView({
      state,
      parent: container
    })
    
    if (height === 'auto' && container) {
      const contentHeight = container.querySelector('.cm-content')?.scrollHeight || 0
      const lineHeight = 21
      const lines = editorView.state.doc.lines
      const newHeight = Math.min(Math.max(contentHeight, lineHeight * 3), lineHeight * 20)
      container.style.height = `${newHeight}px`
    }
    
    const updateDarkMode = () => {
      const newDarkMode = checkDarkMode()
      if (newDarkMode !== darkMode && editorView) {
        darkMode = newDarkMode
        const newTheme = createThemeExtension()
        editorView.dispatch({
          effects: themeCompartment.reconfigure(newTheme)
        })
        forceLinting(editorView)
      }
    }
    
    const observer = new MutationObserver(updateDarkMode)
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class']
    })
    
    return () => {
      observer.disconnect()
      if (editorView) {
        editorView.destroy()
        editorView = null
      }
    }
  })
  
  function updateValue() {
    if (editorView && value !== editorView.state.doc.toString()) {
      const currentCursor = editorView.state.selection.main.head
      editorView.dispatch({
        changes: {
          from: 0,
          to: editorView.state.doc.length,
          insert: value
        },
        selection: { anchor: Math.min(currentCursor, value.length) }
      })
    }
  }
  
  function updateReadOnly() {
    if (editorView) {
      editorView.dispatch({
        effects: readOnlyCompartment.reconfigure(readOnly ? [EditorView.editable.of(false)] : [])
      })
    }
  }
  
  function formatContent() {
    if (!editorView) return
    
    const current = editorView.state.doc.toString()
    const formatted = formatJson(current)
    
    if (formatted !== current) {
      editorView.dispatch({
        changes: {
          from: 0,
          to: editorView.state.doc.length,
          insert: formatted
        }
      })
    }
  }
  
  $: updateValue()
  $: updateReadOnly()
</script>

<div class="relative">
  <div 
    bind:this={container}
    class="rounded-lg border transition-colors {hasErrors ? 'border-red-500 dark:border-red-400' : 'border-gray-300 dark:border-gray-600'}"
    style:min-height="63px"
  ></div>
  
  {#if !readOnly}
    <button
      type="button"
      on:click={formatContent}
      class="absolute top-2 right-2 px-2 py-1 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors opacity-0 group-hover:opacity-100 focus:opacity-100"
      title="Format JSON (Ctrl/Cmd+Shift+F)"
    >
      Format
    </button>
  {/if}
</div>

<style>
  :global(.cm-editor) {
    height: 100%;
  }
  
  :global(.cm-editor:hover .absolute) {
    opacity: 1;
  }
</style>

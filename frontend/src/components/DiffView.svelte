<script>
  export let original = {}
  export let modified = {}
  
  let viewMode = 'split' // 'split' or 'combined'
  
  function flattenObject(obj, prefix = '') {
    const result = {}
    
    for (const key in obj) {
      if (obj.hasOwnProperty(key)) {
        const newKey = prefix ? `${prefix}.${key}` : key
        const value = obj[key]
        
        if (typeof value === 'object' && value !== null && !Array.isArray(value)) {
          Object.assign(result, flattenObject(value, newKey))
        } else {
          result[newKey] = value
        }
      }
    }
    
    return result
  }
  
  function generateDiffLines(original, modified) {
    const flatOrig = flattenObject(original)
    const flatMod = flattenObject(modified)
    
    const allKeys = [...new Set([...Object.keys(flatOrig), ...Object.keys(flatMod)])].sort()
    
    const lines = []
    
    for (const key of allKeys) {
      const origVal = flatOrig[key]
      const modVal = flatMod[key]
      
      const origStr = formatValue(origVal)
      const modStr = formatValue(modVal)
      
      if (origVal === undefined && modVal !== undefined) {
        // Added
        lines.push({
          path: key,
          type: 'added',
          oldLines: [],
          newLines: modStr.split('\n').map(line => ({ content: line, type: 'added' }))
        })
      } else if (origVal !== undefined && modVal === undefined) {
        // Removed
        lines.push({
          path: key,
          type: 'removed',
          oldLines: origStr.split('\n').map(line => ({ content: line, type: 'removed' }))
        })
      } else if (origStr !== modStr) {
        // Modified
        const origLines = origStr.split('\n')
        const modLines = modStr.split('\n')
        
        lines.push({
          path: key,
          type: 'modified',
          oldLines: origLines.map(line => ({ content: line, type: 'removed' })),
          newLines: modLines.map(line => ({ content: line, type: 'added' }))
        })
      }
    }
    
    return lines
  }
  
  function formatValue(value) {
    if (value === undefined) return ''
    if (typeof value === 'string') return `"${value}"`
    if (Array.isArray(value)) {
      if (value.length === 0) return '[]'
      return value.map(v => typeof v === 'string' ? `"${v}"` : v).join(', ')
    }
    return JSON.stringify(value, null, 2)
  }
  
  $: diffLines = generateDiffLines(original, modified)
  $: hasChanges = diffLines.length > 0
  
  function toggleViewMode() {
    viewMode = viewMode === 'split' ? 'combined' : 'split'
  }
</script>

<div class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
  <div class="flex items-center justify-between mb-4">
    <h2 class="text-xl font-semibold text-gray-900 dark:text-white">Metadata Diff</h2>
    <div class="flex items-center gap-2">
      <button
        on:click={toggleViewMode}
        class="px-3 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors text-gray-700 dark:text-gray-300"
      >
        {viewMode === 'split' ? 'Switch to Combined View' : 'Switch to Split View'}
      </button>
    </div>
  </div>
  
  {#if !hasChanges}
    <p class="text-gray-500 dark:text-gray-400 text-sm">No changes detected between original and modified metadata.</p>
  {:else}
    <div class="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
      <!-- Header -->
      <div class="grid grid-cols-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-700">
        <div class="px-4 py-2 text-xs font-medium text-gray-600 dark:text-gray-300 border-r border-gray-200 dark:border-gray-700">
          Original
        </div>
        <div class="px-4 py-2 text-xs font-medium text-gray-600 dark:text-gray-300">
          Modified
        </div>
      </div>
      
      <!-- Diff Content -->
      <div class="font-mono text-sm">
        {#each diffLines as block}
          <!-- Path Header -->
          <div class="bg-gray-100 dark:bg-gray-700 px-4 py-1 text-xs text-gray-600 dark:text-gray-300 border-t border-b border-gray-200 dark:border-gray-600">
            @@ {block.path} @@
          </div>
          
          {#if viewMode === 'split'}
            <!-- Split View -->
            <div class="grid grid-cols-2">
              <div class="border-r border-gray-200 dark:border-gray-700">
                {#if block.oldLines && block.oldLines.length > 0}
                  {#each block.oldLines as line}
                    <div class="px-4 py-0.5 bg-red-50 dark:bg-red-900/20 text-red-800 dark:text-red-300 flex">
                      <span class="w-6 flex-shrink-0 select-none">-</span>
                      <span class="whitespace-pre-wrap break-all">{line.content}</span>
                    </div>
                  {/each}
                {:else}
                  <div class="px-4 py-2 text-gray-400 dark:text-gray-500 italic">no change</div>
                {/if}
              </div>
              <div>
                {#if block.newLines && block.newLines.length > 0}
                  {#each block.newLines as line}
                    <div class="px-4 py-0.5 bg-green-50 dark:bg-green-900/20 text-green-800 dark:text-green-300 flex">
                      <span class="w-6 flex-shrink-0 select-none">+</span>
                      <span class="whitespace-pre-wrap break-all">{line.content}</span>
                    </div>
                  {/each}
                {:else}
                  <div class="px-4 py-2 text-gray-400 dark:text-gray-500 italic">no change</div>
                {/if}
              </div>
            </div>
          {:else}
            <!-- Combined/Unified View -->
            <div>
              {#if block.oldLines && block.oldLines.length > 0}
                {#each block.oldLines as line}
                  <div class="px-4 py-0.5 bg-red-50 dark:bg-red-900/20 text-red-800 dark:text-red-300 flex">
                    <span class="w-6 flex-shrink-0 select-none">-</span>
                    <span class="whitespace-pre-wrap break-all">{line.content}</span>
                  </div>
                {/each}
              {/if}
              {#if block.newLines && block.newLines.length > 0}
                {#each block.newLines as line}
                  <div class="px-4 py-0.5 bg-green-50 dark:bg-green-900/20 text-green-800 dark:text-green-300 flex">
                    <span class="w-6 flex-shrink-0 select-none">+</span>
                    <span class="whitespace-pre-wrap break-all">{line.content}</span>
                  </div>
                {/each}
              {/if}
            </div>
          {/if}
        {/each}
      </div>
    </div>
    
    <!-- Legend -->
    <div class="mt-4 flex items-center gap-4 text-xs text-gray-600 dark:text-gray-400">
      <div class="flex items-center gap-2">
        <span class="w-4 h-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 inline-block"></span>
        <span>Removed</span>
      </div>
      <div class="flex items-center gap-2">
        <span class="w-4 h-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 inline-block"></span>
        <span>Added</span>
      </div>
    </div>
  {/if}
</div>

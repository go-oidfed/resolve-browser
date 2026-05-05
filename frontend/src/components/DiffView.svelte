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

<div class="bg-white rounded-lg shadow-md p-6">
  <div class="flex items-center justify-between mb-4">
    <h2 class="text-xl font-semibold text-gray-900">Metadata Diff</h2>
    <div class="flex items-center gap-2">
      <button
        on:click={toggleViewMode}
        class="px-3 py-1 text-sm border rounded-md hover:bg-gray-50 transition-colors"
      >
        {viewMode === 'split' ? 'Switch to Combined View' : 'Switch to Split View'}
      </button>
    </div>
  </div>
  
  {#if !hasChanges}
    <p class="text-gray-500 text-sm">No changes detected between original and modified metadata.</p>
  {:else}
    <div class="border rounded-lg overflow-hidden">
      <!-- Header -->
      <div class="grid grid-cols-2 border-b bg-gray-50">
        <div class="px-4 py-2 text-xs font-medium text-gray-600 border-r">
          Original
        </div>
        <div class="px-4 py-2 text-xs font-medium text-gray-600">
          Modified
        </div>
      </div>
      
      <!-- Diff Content -->
      <div class="font-mono text-sm">
        {#each diffLines as block}
          <!-- Path Header -->
          <div class="bg-gray-100 px-4 py-1 text-xs text-gray-600 border-t border-b">
            @@ {block.path} @@
          </div>
          
          {#if viewMode === 'split'}
            <!-- Split View -->
            <div class="grid grid-cols-2">
              <div class="border-r">
                {#if block.oldLines && block.oldLines.length > 0}
                  {#each block.oldLines as line}
                    <div class="px-4 py-0.5 bg-red-50 text-red-800 flex">
                      <span class="w-6 flex-shrink-0 select-none">-</span>
                      <span class="whitespace-pre-wrap break-all">{line.content}</span>
                    </div>
                  {/each}
                {:else}
                  <div class="px-4 py-2 text-gray-400 italic">no change</div>
                {/if}
              </div>
              <div>
                {#if block.newLines && block.newLines.length > 0}
                  {#each block.newLines as line}
                    <div class="px-4 py-0.5 bg-green-50 text-green-800 flex">
                      <span class="w-6 flex-shrink-0 select-none">+</span>
                      <span class="whitespace-pre-wrap break-all">{line.content}</span>
                    </div>
                  {/each}
                {:else}
                  <div class="px-4 py-2 text-gray-400 italic">no change</div>
                {/if}
              </div>
            </div>
          {:else}
            <!-- Combined/Unified View -->
            <div>
              {#if block.oldLines && block.oldLines.length > 0}
                {#each block.oldLines as line}
                  <div class="px-4 py-0.5 bg-red-50 text-red-800 flex">
                    <span class="w-6 flex-shrink-0 select-none">-</span>
                    <span class="whitespace-pre-wrap break-all">{line.content}</span>
                  </div>
                {/each}
              {/if}
              {#if block.newLines && block.newLines.length > 0}
                {#each block.newLines as line}
                  <div class="px-4 py-0.5 bg-green-50 text-green-800 flex">
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
    <div class="mt-4 flex items-center gap-4 text-xs text-gray-600">
      <div class="flex items-center gap-2">
        <span class="w-4 h-4 bg-red-50 border border-red-200 inline-block"></span>
        <span>Removed</span>
      </div>
      <div class="flex items-center gap-2">
        <span class="w-4 h-4 bg-green-50 border border-green-200 inline-block"></span>
        <span>Added</span>
      </div>
    </div>
  {/if}
</div>

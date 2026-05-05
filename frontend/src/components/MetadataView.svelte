<script>
  import { copyToClipboard } from '../lib/utils.js'
  
  export let metadata = {}
  export let title = 'Metadata'
  
  let expandedSections = new Set()
  let copied = false
  
  function toggleSection(type) {
    if (expandedSections.has(type)) {
      expandedSections.delete(type)
    } else {
      expandedSections.add(type)
    }
    expandedSections = new Set(expandedSections)
  }
  
  function expandAll() {
    expandedSections = new Set(Object.keys(metadata))
  }
  
  function collapseAll() {
    expandedSections = new Set()
  }
  
  function formatType(type) {
    return type.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())
  }
  
  function copyMetadata() {
    copyToClipboard(JSON.stringify(metadata, null, 2))
    copied = true
    setTimeout(() => copied = false, 2000)
  }
</script>

<div class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
  <div class="flex items-center justify-between mb-4">
    <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{title}</h2>
    <div class="flex gap-2">
      <button
        on:click={expandAll}
        class="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 font-medium"
      >
        Expand All
      </button>
      <span class="text-gray-300 dark:text-gray-600">|</span>
      <button
        on:click={collapseAll}
        class="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 font-medium"
      >
        Collapse All
      </button>
      <span class="text-gray-300 dark:text-gray-600">|</span>
      <button
        on:click={copyMetadata}
        class="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 font-medium"
      >
        {copied ? 'Copied!' : 'Copy'}
      </button>
    </div>
  </div>
  
  {#if Object.keys(metadata).length === 0}
    <p class="text-gray-500 dark:text-gray-400 text-sm">No metadata available</p>
  {:else}
    <div class="space-y-3">
      {#each Object.entries(metadata) as [type, data]}
        <div class="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
          <button
            on:click={() => toggleSection(type)}
            class="w-full flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700 hover:bg-gray-100 dark:hover:bg-gray-600 transition-colors"
          >
            <span class="font-medium text-gray-900 dark:text-white">{formatType(type)}</span>
            <svg
              class="w-5 h-5 text-gray-500 dark:text-gray-400 transform transition-transform {expandedSections.has(type) ? 'rotate-180' : ''}"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
            </svg>
          </button>
          
          {#if expandedSections.has(type)}
            <div class="p-3 bg-white dark:bg-gray-800">
              <pre class="text-xs bg-gray-50 dark:bg-gray-900 p-3 rounded overflow-x-auto max-h-64 overflow-y-auto"><code class="text-gray-900 dark:text-gray-100">{JSON.stringify(data, null, 2)}</code></pre>
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>

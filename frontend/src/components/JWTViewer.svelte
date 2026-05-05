<script>
  import { copyToClipboard } from '../lib/utils.js'
  
  export let header = {}
  export let payload = {}
  export let rawJwt = ''
  
  let activeTab = 'payload'
  let copied = false
  
  function copyRaw() {
    copyToClipboard(rawJwt)
    copied = true
    setTimeout(() => copied = false, 2000)
  }
</script>

<div class="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
  <div class="flex border-b border-gray-200 dark:border-gray-700 bg-gray-100 dark:bg-gray-700">
    <button
      on:click={() => activeTab = 'header'}
      class="px-4 py-2 text-sm font-medium transition-colors {
        activeTab === 'header'
          ? 'bg-white dark:bg-gray-800 text-blue-600 dark:text-blue-400 border-b-2 border-blue-600 dark:border-blue-400'
          : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'
      }"
    >
      Header
    </button>
    <button
      on:click={() => activeTab = 'payload'}
      class="px-4 py-2 text-sm font-medium transition-colors {
        activeTab === 'payload'
          ? 'bg-white dark:bg-gray-800 text-blue-600 dark:text-blue-400 border-b-2 border-blue-600 dark:border-blue-400'
          : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'
      }"
    >
      Payload
    </button>
    <button
      on:click={() => activeTab = 'raw'}
      class="px-4 py-2 text-sm font-medium transition-colors {
        activeTab === 'raw'
          ? 'bg-white dark:bg-gray-800 text-blue-600 dark:text-blue-400 border-b-2 border-blue-600 dark:border-blue-400'
          : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'
      }"
    >
      Raw JWT
    </button>
  </div>
  
  <div class="p-4 bg-white dark:bg-gray-800">
    {#if activeTab === 'header'}
      <pre class="text-xs bg-gray-50 dark:bg-gray-900 p-3 rounded overflow-x-auto"><code class="text-gray-900 dark:text-gray-100">{JSON.stringify(header, null, 2)}</code></pre>
    {:else if activeTab === 'payload'}
      <pre class="text-xs bg-gray-50 dark:bg-gray-900 p-3 rounded overflow-x-auto max-h-96 overflow-y-auto"><code class="text-gray-900 dark:text-gray-100">{JSON.stringify(payload, null, 2)}</code></pre>
    {:else if activeTab === 'raw'}
      <div class="relative">
        <pre class="text-xs bg-gray-50 dark:bg-gray-900 p-3 rounded overflow-x-auto break-all pr-20 text-gray-900 dark:text-gray-100">{rawJwt}</pre>
        <button
          on:click={copyRaw}
          class="absolute top-2 right-2 px-3 py-1 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
        >
          {copied ? 'Copied!' : 'Copy'}
        </button>
      </div>
    {/if}
  </div>
</div>

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

<div class="border rounded-lg overflow-hidden">
  <div class="flex border-b bg-gray-100">
    <button
      on:click={() => activeTab = 'header'}
      class="px-4 py-2 text-sm font-medium transition-colors {
        activeTab === 'header'
          ? 'bg-white text-blue-600 border-b-2 border-blue-600'
          : 'text-gray-600 hover:text-gray-900'
      }"
    >
      Header
    </button>
    <button
      on:click={() => activeTab = 'payload'}
      class="px-4 py-2 text-sm font-medium transition-colors {
        activeTab === 'payload'
          ? 'bg-white text-blue-600 border-b-2 border-blue-600'
          : 'text-gray-600 hover:text-gray-900'
      }"
    >
      Payload
    </button>
    <button
      on:click={() => activeTab = 'raw'}
      class="px-4 py-2 text-sm font-medium transition-colors {
        activeTab === 'raw'
          ? 'bg-white text-blue-600 border-b-2 border-blue-600'
          : 'text-gray-600 hover:text-gray-900'
      }"
    >
      Raw JWT
    </button>
  </div>
  
  <div class="p-4 bg-white">
    {#if activeTab === 'header'}
      <pre class="text-xs bg-gray-50 p-3 rounded overflow-x-auto"><code>{JSON.stringify(header, null, 2)}</code></pre>
    {:else if activeTab === 'payload'}
      <pre class="text-xs bg-gray-50 p-3 rounded overflow-x-auto max-h-96 overflow-y-auto"><code>{JSON.stringify(payload, null, 2)}</code></pre>
    {:else if activeTab === 'raw'}
      <div class="relative">
        <pre class="text-xs bg-gray-50 p-3 rounded overflow-x-auto break-all pr-20">{rawJwt}</pre>
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

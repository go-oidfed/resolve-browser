<script>
  import StatementCard from './StatementCard.svelte'
  
  export let chain = []
  export let editedChain = []
  export let onEditChange
  
  let expandedCards = new Set()
  
  function toggleCard(index) {
    if (expandedCards.has(index)) {
      expandedCards.delete(index)
    } else {
      expandedCards.add(index)
    }
    expandedCards = new Set(expandedCards)
  }
  
  function expandAll() {
    expandedCards = new Set(chain.map((_, i) => i))
  }
  
  function collapseAll() {
    expandedCards = new Set()
  }
</script>

<div class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
  <div class="flex items-center justify-between mb-4">
    <h2 class="text-xl font-semibold text-gray-900 dark:text-white">Trust Chain</h2>
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
    </div>
  </div>
  
  <div class="space-y-0">
    {#each chain as statement, index}
      <div class="relative">
        <StatementCard
          {statement}
          index={index}
          isExpanded={expandedCards.has(index)}
          onToggle={() => toggleCard(index)}
          editedStatement={editedChain[index]}
          onEditChange={(field, value) => onEditChange(index, field, value)}
        />
        
        {#if index < chain.length - 1}
          <div class="flex justify-center py-2">
            <div class="w-0.5 h-8 bg-gray-300 dark:bg-gray-600"></div>
          </div>
        {/if}
      </div>
    {/each}
  </div>
</div>

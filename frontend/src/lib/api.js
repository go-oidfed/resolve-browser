export async function resolveTrustChain(data) {
  const response = await fetch('/api/resolve', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  })
  
  if (!response.ok) {
    const errorData = await response.json()
    throw new Error(errorData.error_description || errorData.error || 'Request failed')
  }
  
  return await response.json()
}

export async function previewEditedChain(trustChain) {
  const response = await fetch('/api/resolve/preview', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(trustChain),
  })
  
  const responseData = await response.json()
  
  if (!response.ok) {
    return {
      error: responseData.error || 'preview_failed',
      error_description: responseData.error_description || 'Unknown error occurred'
    }
  }
  
  return responseData
}

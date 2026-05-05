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

export async function previewEditedChain(data) {
  const response = await fetch('/api/resolve/preview', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  })
  
  const responseData = await response.json()
  
  // Return error data in the response instead of throwing
  // This allows the UI to display the error nicely
  if (!response.ok) {
    return {
      error: responseData.error || 'preview_failed',
      error_description: responseData.error_description || 'Unknown error occurred'
    }
  }
  
  return responseData
}

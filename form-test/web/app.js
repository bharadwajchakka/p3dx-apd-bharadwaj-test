document.addEventListener('DOMContentLoaded', function(){
  const dataBtn = document.getElementById('dataOwnerBtn')
  const outBtn = document.getElementById('outputOwnerBtn')
  const dataSection = document.getElementById('dataFormSection')
  const outSection = document.getElementById('outputFormSection')
  const dataForm = document.getElementById('dataForm')
  const outForm = document.getElementById('outputForm')
  const message = document.getElementById('message')

  // Defensive checks to avoid JS errors if an element is missing
  if (!dataBtn || !outBtn || !dataSection || !outSection || !message) {
    console.error('UI initialization failed: missing element(s)', {dataBtn, outBtn, dataSection, outSection, message})
    return
  }
  const messageJson = document.getElementById('messageJson')
  const messageClose = document.getElementById('messageClose')
  const messageTitle = document.getElementById('messageTitle')

  if (!messageJson || !messageClose || !messageTitle) {
    console.error('message sub-elements missing', {messageJson, messageClose, messageTitle})
    return
  }

  dataBtn.addEventListener('click', ()=>{
    console.debug('Data Owner button clicked')
    dataSection.classList.remove('hidden')
    outSection.classList.add('hidden')
  })
  outBtn.addEventListener('click', ()=>{
    console.debug('Output Owner button clicked')
    outSection.classList.remove('hidden')
    dataSection.classList.add('hidden')
  })

  function showMessage(text){
    message.textContent = text
    message.classList.remove('hidden')
  }

  function submitFormPayload(payload, kind){
    fetch('http://localhost:8080/api/v1/form-submissions', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ payload })
    })
      .then(res => res.json())
      .then(data => {
        if (data.status === 'success') {
          console.debug(kind + ' payload saved to APD', data)
        } else {
          console.error('failed to save ' + kind + ' payload', data)
          showMessage(kind + ' payload saved locally, but APD storage failed: ' + (data.message || 'unknown error'))
        }
      })
      .catch(err => {
        console.error('error saving ' + kind + ' payload', err)
        showMessage(kind + ' payload saved locally, but APD storage failed: ' + err.message)
      })
  }

  dataForm.addEventListener('submit', function(e){
    e.preventDefault()
    const fd = new FormData(dataForm)
    const obj = {
      form_id: fd.get('form_id'),
      // requested_by: fd.get('requested_by'),
      data_owner_id: fd.get('data_owner_id'),
      ram: Number(fd.get('RAM')),
      // num_cpus: Number(fd.get('num_cpus')),
      // num_gpus: Number(fd.get('num_gpus')),
      memory_mb: Number(fd.get('memory_mb')),
      data_size_bytes: fd.get('data_size_bytes') ? Number(fd.get('data_size_bytes')) : undefined,
      data_resource_id: fd.get('data_resource_id') || undefined,
      filled: true,
      // requested_at: new Date().toISOString(),
      filled_at: new Date().toISOString()
    }
    // store form payload in APD and show user the submitted form as JSON
    submitFormPayload(obj, 'Data Owner')
    dataSection.classList.add('hidden')
    messageTitle.textContent = 'Data Owner Form JSON'
    messageJson.textContent = JSON.stringify(obj, null, 2)
    message.classList.remove('hidden')
  })

  outForm.addEventListener('submit', function(e){
    e.preventDefault()
    const fd = new FormData(outForm)
    const compRaw = fd.get('components') || ''
    const comps = {}
    compRaw.split(',').map(s=>s.trim()).filter(Boolean).forEach(pair=>{
      const kv = pair.split('=')
      if(kv.length===2) comps[kv[0].trim()] = kv[1].trim()
    })
    const obj = {
      form_id: fd.get('form_id'),
      requested_by: fd.get('requested_by'),
      output_owner_id: fd.get('output_owner_id'),
      num_server_rounds: Number(fd.get('num_server_rounds')),
      fraction_evaluate: Number(fd.get('fraction_evaluate')),
      local_epochs: Number(fd.get('local_epochs')),
      learning_rate: Number(fd.get('learning_rate')),
      batch_size: Number(fd.get('batch_size')),
      model: fd.get('model') || undefined,
      framework: fd.get('framework') || undefined,
      components: Object.keys(comps).length?comps:undefined,
      filled: true,
      requested_at: new Date().toISOString(),
      filled_at: new Date().toISOString()
    }
    // store form payload in APD and show user the submitted form as JSON
    submitFormPayload(obj, 'Output Owner')
    outSection.classList.add('hidden')
    messageTitle.textContent = 'Output Owner Form JSON'
    messageJson.textContent = JSON.stringify(obj, null, 2)
    message.classList.remove('hidden')
  })

  messageClose.addEventListener('click', ()=>{
    message.classList.add('hidden')
    messageJson.textContent = ''
  })

})

<script setup lang="ts">
import { ref, onMounted, nextTick, computed, watch } from 'vue'
import MarkdownIt from 'markdown-it'
import request from '@/api/request'

const md = new MarkdownIt({ html: true, breaks: true, linkify: true })

function renderMarkdown(text: string): string {
  if (!text) return ''
  try {
    return md.render(text)
  } catch {
    return text
  }
}

interface Message {
  role: 'user' | 'assistant'
  content: string
  sources?: { title: string; content: string; score: number }[]
  images?: string[]
  model?: string
  loading?: boolean
}

interface PlatformPreset {
  name: string
  type: string
  api_key: string
  base_url: string
  enabled: boolean
  models: string[]
}

const messages = ref<Message[]>([])
const inputText = ref('')
const sending = ref(false)
const chatBody = ref<HTMLElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const pendingImages = ref<string[]>([])
const deepThinking = ref(false)
const platforms = ref<PlatformPreset[]>([])
const platformsLoading = ref(false)
const selectedPlatform = ref('')
const selectedModel = ref('')
const selectedTool = ref('search')

interface ToolOption {
  label: string
  value: string
  key?: string
}
const toolOptions = ref<ToolOption[]>([])

const toolOptionsDisplay = computed(() => {
  const kb: ToolOption = { label: '知识库', value: 'search' }
  const tools = toolOptions.value.filter(t => t.value !== 'search')
  return [kb, ...tools]
})

const modelsForPlatform = computed(() => {
  const p = platforms.value.find(x => x.name === selectedPlatform.value)
  return p?.models || []
})

function selectFirstModel() {
  const models = modelsForPlatform.value
  if (models.length > 0) {
    selectedModel.value = models[0]
  }
}

const canSend = computed(() => (inputText.value.trim() !== '' || pendingImages.value.length > 0) && selectedModel.value !== '')

watch(selectedPlatform, () => {
  selectedModel.value = ''
  selectFirstModel()
})

async function fetchPlatforms() {
  platformsLoading.value = true
  try {
    const res: any = await request.get('/system-config', { params: { page: 1, page_size: 200 } })
    const list: any[] = res.data?.data || []
    platforms.value = list
      .filter(c => c.category === 'llm' && c.key.startsWith('preset:'))
      .map(c => {
        try { return JSON.parse(c.value) } catch { return null }
      })
      .filter(Boolean) as PlatformPreset[]
    if (platforms.value.length > 0 && !selectedPlatform.value) {
      selectedPlatform.value = platforms.value[0].name
      selectFirstModel()
    }
  } catch {
    platforms.value = []
  } finally {
    platformsLoading.value = false
  }
}

function scrollToBottom() {
  nextTick(() => {
    if (chatBody.value) {
      chatBody.value.scrollTop = chatBody.value.scrollHeight
    }
  })
}

function pickImage() {
  fileInput.value?.click()
}

function handleImagePick(e: Event) {
  const input = e.target as HTMLInputElement
  if (!input.files?.length) return
  for (const file of input.files) {
    const reader = new FileReader()
    reader.onload = (ev) => {
      const dataUrl = ev.target?.result as string
      if (dataUrl) pendingImages.value.push(dataUrl)
    }
    reader.readAsDataURL(file)
  }
  input.value = ''
}

function removePendingImage(idx: number) {
  pendingImages.value.splice(idx, 1)
}

async function sendMessage() {
  const text = inputText.value.trim()
  if ((!text && pendingImages.value.length === 0) || sending.value) return

  const userImages = [...pendingImages.value]
  pendingImages.value = []
  inputText.value = ''
  nextTick(autoResize)

  messages.value.push({
    role: 'user',
    content: text || '(图片消息)',
    images: userImages.length > 0 ? userImages : undefined,
  })

  const assistantIdx = messages.value.push({ role: 'assistant', content: '', loading: true })
  sending.value = true
  scrollToBottom()

  try {
    const payload: any = { question: text || '请描述这张图片', tool: selectedTool.value }
    if (selectedModel.value) payload.model = selectedModel.value
    if (deepThinking.value) payload.deep_thinking = true
    if (userImages.length > 0) payload.images = userImages

    const baseURL = import.meta.env.VITE_API_BASE_URL || '/api'
    const token = localStorage.getItem('token')
    const resp = await fetch(`${baseURL}/knowledge/ask/stream`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: token ? `Bearer ${token}` : '',
      },
      body: JSON.stringify(payload),
    })
    if (!resp.ok) throw new Error(`HTTP ${resp.status}`)

    const reader = resp.body?.getReader()
    if (!reader) throw new Error('stream not supported')

    const decoder = new TextDecoder()
    let buf = ''
    let currentEvent = ''
    let dataLines: string[] = []
    let sources: any[] = []

    function flushEvent() {
      if (dataLines.length === 0) return
      const data = dataLines.join('\n')
      dataLines = []
      if (currentEvent === 'error') {
        messages.value[assistantIdx - 1].content = data
      } else if (currentEvent === 'sources') {
        try { sources = JSON.parse(data) } catch {}
      } else if (currentEvent === 'done') {
        // stream complete
      } else {
        messages.value[assistantIdx - 1].content += data
        scrollToBottom()
      }
    }

    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      buf += decoder.decode(value, { stream: true })

      while (true) {
        const sep = buf.indexOf('\n')
        if (sep === -1) break
        const line = buf.slice(0, sep)
        buf = buf.slice(sep + 1)

        if (line.startsWith('event: ')) {
          flushEvent()
          currentEvent = line.slice(7).trim()
        } else if (line.startsWith('data: ')) {
          dataLines.push(line.slice(6))
        } else if (line === '') {
          flushEvent()
        } else if (dataLines.length > 0) {
          dataLines[dataLines.length - 1] += '\n' + line
        }
      }
    }
    flushEvent()
    if (sources.length) {
      messages.value[assistantIdx - 1].sources = sources
    }
  } catch (e: any) {
    if (!messages.value[assistantIdx - 1].content) {
      messages.value[assistantIdx - 1].content = e?.message || '查询失败，请稍后重试'
    }
  } finally {
    messages.value[assistantIdx - 1].loading = false
    sending.value = false
    scrollToBottom()
  }
}

function handleKeydown(e: Event | KeyboardEvent) {
  const ev = e as KeyboardEvent
  if (ev.key === 'Enter' && !ev.shiftKey) {
    ev.preventDefault()
    sendMessage()
  }
}

const textareaRef = ref<HTMLTextAreaElement | null>(null)
const copiedIdx = ref(-1)

function autoResize() {
  const el = textareaRef.value
  if (!el) return
  el.style.height = ''
  el.style.height = el.scrollHeight + 'px'
}

async function copyContent(text: string, idx: number) {
  try {
    await navigator.clipboard.writeText(text)
    copiedIdx.value = idx
    setTimeout(() => { if (copiedIdx.value === idx) copiedIdx.value = -1 }, 2000)
  } catch {
    // fallback
    const ta = document.createElement('textarea')
    ta.value = text
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    document.body.removeChild(ta)
    copiedIdx.value = idx
    setTimeout(() => { if (copiedIdx.value === idx) copiedIdx.value = -1 }, 2000)
  }
}

async function fetchTools() {
  try {
    const res: any = await request.get('/mcps', { params: { page: 1, page_size: 200 } })
    const items: any[] = res?.list || res?.data?.list || res?.data?.data || res?.data || []
    toolOptions.value = items
      .filter((t: any) => t.enabled !== false)
      .map((t: any) => ({ label: t.name, value: t.key || t.name, key: t.key }))
  } catch {
    // fallback to defaults
    toolOptions.value = []
  }
}

onMounted(() => {
  fetchPlatforms()
  fetchTools()
})
</script>

<template>
  <div class="chat-page">

    <div v-if="messages.length === 0" class="chat-empty">
      <div class="empty-icon">✨</div>
      <h2>有什么可以帮助你的？</h2>
      <p>基于知识库和资料文档，向我提问</p>
    </div>

    <div v-else ref="chatBody" class="chat-body">
      <div v-for="(msg, idx) in messages" :key="idx" class="msg-row" :class="msg.role">
        <div class="msg-avatar">{{ msg.role === 'user' ? 'U' : 'A' }}</div>
        <div class="msg-main">
          <div class="msg-bubble">
            <div v-if="msg.images?.length" class="msg-imgs">
              <img v-for="(img, ii) in msg.images" :key="ii" :src="img" />
            </div>
            <div class="msg-text" v-html="renderMarkdown(msg.content)"></div>
            <div v-if="msg.loading" class="msg-status generating">
              <span class="status-dot" /><span class="status-dot" /><span class="status-dot" />
              <span class="status-text">正在生成回答</span>
            </div>
            <div v-else-if="msg.role === 'assistant' && msg.content" class="msg-status done">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" style="color:#2468f2;"><polyline points="20 6 9 17 4 12"/></svg>
              <span class="status-text">回答完成</span>
              <button class="copy-btn" title="复制内容" @click="copyContent(msg.content, idx)">
                <template v-if="copiedIdx === idx">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
                </template>
                <template v-else>
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
                </template>
              </button>
            </div>
          </div>
          <div v-if="msg.sources?.length" class="msg-sources">
            <div class="sources-title">参考来源 ({{ msg.sources.length }})</div>
            <div v-for="(s, si) in msg.sources" :key="si" class="source-item">
              <strong>{{ s.title }}</strong>
              <p>{{ s.content?.slice(0, 160) }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="chat-footer">
      <div class="input-panel">
        <div v-if="pendingImages.length" class="preview-bar">
          <div v-for="(img, idx) in pendingImages" :key="idx" class="preview-item">
            <img :src="img" />
            <button class="preview-close" @click="removePendingImage(idx)">×</button>
          </div>
        </div>
        <div class="input-row">
          <div class="text-input-wrap">
            <textarea
              ref="textareaRef"
              v-model="inputText"
              placeholder="输入你的问题..."
              :disabled="sending"
              rows="1"
              @keydown="handleKeydown"
              @input="autoResize"
            ></textarea>
          </div>
        </div>
        <div class="input-tools">
          <div class="tools-left">
            <button class="tool-chip tool-plus" title="上传图片" @click="pickImage">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            </button>
            <input ref="fileInput" type="file" accept="image/*" multiple style="display:none" @change="handleImagePick" />
            <button class="tool-chip" :class="{ active: deepThinking }" @click="deepThinking = !deepThinking">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="16 3 21 3 21 8"/><line x1="4" y1="20" x2="21" y2="3"/><polyline points="21 16 21 21 16 21"/><line x1="15" y1="15" x2="21" y2="21"/><line x1="4" y1="4" x2="9" y2="9"/></svg>
              深度思考
            </button>
            <el-select
              v-model="selectedTool"
              size="small"
              style="width:120px"
            >
              <el-option v-for="opt in toolOptionsDisplay" :key="opt.value" :label="opt.label" :value="opt.value" />
            </el-select>
          </div>
          <div class="tools-right">
            <el-select
              v-model="selectedPlatform"
              placeholder="平台"
              size="small"
              :loading="platformsLoading"
              style="width:120px"
              clearable
            >
              <el-option v-for="p in platforms" :key="p.name" :label="p.name" :value="p.name">
                <span>{{ p.name }}</span>
                <el-tag size="small" type="info" style="margin-left:6px">{{ p.type }}</el-tag>
              </el-option>
            </el-select>
            <el-select
              v-model="selectedModel"
              placeholder="模型"
              size="small"
              style="width:150px"
              :disabled="!selectedPlatform"
              clearable
            >
              <el-option v-for="m in modelsForPlatform" :key="m" :label="m" :value="m" />
            </el-select>
            <button class="tool-chip send-btn" :class="{ active: canSend }" :disabled="!canSend || sending" @click="sendMessage">
              <template v-if="!sending">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>
              </template>
              <span v-else class="sending-dots">...</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.chat-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 120px);
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
}

.chat-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #8c8c8c;
  padding-bottom: 80px;
}
.empty-icon {
  font-size: 48px;
  margin-bottom: 12px;
}
.chat-empty h2 {
  margin: 0 0 6px;
  color: #1d2129;
  font-size: 20px;
  font-weight: 600;
}
.chat-empty p {
  margin: 0;
  font-size: 14px;
  color: #86909c;
}

.chat-body {
  flex: 1;
  overflow-y: auto;
  padding: 24px 20px;
}
.msg-row {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}
.msg-row.user {
  flex-direction: row-reverse;
}
.msg-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 13px;
  flex-shrink: 0;
}
.msg-row.user .msg-avatar {
  background: linear-gradient(135deg, #2468f2, #4e8aff);
  color: #fff;
}
.msg-row.assistant .msg-avatar {
  background: linear-gradient(135deg, #00b96b, #34d687);
  color: #fff;
}
.msg-main {
  max-width: 72%;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.msg-bubble {
  padding: 12px 16px;
  border-radius: 14px;
  line-height: 1.7;
  font-size: 14px;
  word-break: break-word;
}
.msg-row.user .msg-bubble {
  background: #2468f2;
  color: #fff;
  border-bottom-right-radius: 4px;
}
.msg-row.assistant .msg-bubble {
  background: #f2f3f5;
  color: #1d2129;
  border-bottom-left-radius: 4px;
}
.msg-imgs {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 8px;
}
.msg-imgs img {
  max-width: 180px;
  max-height: 180px;
  border-radius: 8px;
  object-fit: cover;
}
.msg-text {
  line-height: 1.7;
  word-wrap: break-word;
}
.msg-streaming {
  white-space: pre-wrap;
}
.cursor {
  animation: cursorBlink 0.8s step-end infinite;
  color: #2468f2;
  font-size: 14px;
  margin-left: 2px;
}
@keyframes cursorBlink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}
.msg-text :deep(p) { margin: 0 0 8px; }
.msg-text :deep(p:last-child) { margin-bottom: 0; }
.msg-text :deep(code) {
  background: #f0f0f0;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 13px;
  font-family: 'SF Mono', 'Cascadia Code', monospace;
}
.msg-text :deep(pre) {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 12px 16px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 8px 0;
  font-size: 13px;
  line-height: 1.5;
}
.msg-text :deep(pre code) {
  background: none;
  padding: 0;
  color: inherit;
}
.msg-text :deep(ul), .msg-text :deep(ol) {
  padding-left: 20px;
  margin: 4px 0;
}
.msg-text :deep(li) { margin: 2px 0; }
.msg-text :deep(h1), .msg-text :deep(h2), .msg-text :deep(h3),
.msg-text :deep(h4), .msg-text :deep(h5), .msg-text :deep(h6) {
  margin: 12px 0 6px;
  font-weight: 600;
  color: #1d2129;
}
.msg-text :deep(h1) { font-size: 18px; }
.msg-text :deep(h2) { font-size: 16px; }
.msg-text :deep(h3) { font-size: 15px; }
.msg-text :deep(blockquote) {
  border-left: 3px solid #2468f2;
  margin: 8px 0;
  padding: 4px 12px;
  color: #666;
  background: #f7f8fa;
  border-radius: 0 6px 6px 0;
}
.msg-text :deep(table) {
  border-collapse: collapse;
  width: 100%;
  margin: 8px 0;
  font-size: 13px;
}
.msg-text :deep(th), .msg-text :deep(td) {
  border: 1px solid #e5e6eb;
  padding: 6px 10px;
  text-align: left;
}
.msg-text :deep(th) {
  background: #f7f8fa;
  font-weight: 600;
}
.msg-text :deep(a) { color: #2468f2; text-decoration: none; }
.msg-text :deep(a:hover) { text-decoration: underline; }
.msg-text :deep(hr) {
  border: none;
  border-top: 1px solid #e5e6eb;
  margin: 12px 0;
}
.msg-text :deep(img) {
  max-width: 100%;
  border-radius: 6px;
  margin: 8px 0;
}
.msg-status {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid #e5e6eb;
  font-size: 12px;
}
.msg-status .status-text {
  color: #86909c;
}
.msg-status.generating .status-text {
  color: #2468f2;
}
.msg-status.done .status-text {
  color: #86909c;
}
.copy-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-left: auto;
  width: 26px;
  height: 26px;
  border: 1px solid transparent;
  border-radius: 6px;
  background: transparent;
  color: #86909c;
  cursor: pointer;
  transition: all 0.15s;
}
.copy-btn:hover {
  background: #f2f3f5;
  border-color: #e5e6eb;
  color: #4e5969;
}
.status-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: #2468f2;
  animation: statusPulse 1.2s ease-in-out infinite;
}
.status-dot:nth-child(2) { animation-delay: 0.2s; }
.status-dot:nth-child(3) { animation-delay: 0.4s; }
@keyframes statusPulse {
  0%, 80%, 100% { opacity: 0.3; transform: scale(0.8); }
  40% { opacity: 1; transform: scale(1.1); }
}

.msg-sources {
  font-size: 13px;
}
.sources-title {
  color: #86909c;
  margin-bottom: 4px;
  cursor: pointer;
  font-size: 12px;
}
.source-item {
  background: #f7f8fa;
  border-radius: 8px;
  padding: 8px 12px;
  margin-bottom: 6px;
  border: 1px solid #e8e8e8;
}
.source-item strong {
  display: block;
  margin-bottom: 2px;
  font-size: 13px;
}
.source-item p {
  margin: 0;
  color: #666;
  font-size: 12px;
  white-space: pre-wrap;
}

.chat-footer {
  padding: 0 20px 20px;
  flex-shrink: 0;
}
.input-panel {
  border: 1px solid #e5e6eb;
  border-radius: 14px;
  background: #fff;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);
  transition: border-color 0.2s, box-shadow 0.2s;
}
.input-panel:focus-within {
  border-color: #2468f2;
  box-shadow: 0 0 0 3px rgba(36,104,242,0.08);
}
.preview-bar {
  display: flex;
  gap: 8px;
  padding: 12px 12px 0;
  flex-wrap: wrap;
}
.preview-item {
  position: relative;
  width: 56px;
  height: 56px;
}
.preview-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 8px;
  border: 1px solid #e5e6eb;
}
.preview-close {
  position: absolute;
  top: -6px;
  right: -6px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  border: none;
  background: #c9cdd4;
  color: #fff;
  font-size: 12px;
  line-height: 18px;
  text-align: center;
  cursor: pointer;
  padding: 0;
}
.preview-close:hover {
  background: #86909c;
}
.input-row {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  padding: 10px 12px 8px;
}
.text-input-wrap {
  flex: 1;
}
.text-input-wrap textarea {
  width: 100%;
  border: none;
  outline: none;
  resize: none;
  font-size: 14px;
  line-height: 1.6;
  color: #1d2129;
  font-family: inherit;
  max-height: 120px;
  background: transparent;
}
.text-input-wrap textarea::placeholder {
  color: #c9cdd4;
}
.sending-dots {
  font-weight: 700;
  font-size: 16px;
  letter-spacing: 2px;
}
.input-tools {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 10px 10px;
}
.tools-left {
  display: flex;
  align-items: center;
  gap: 6px;
}
.tools-right {
  display: flex;
  align-items: center;
  gap: 6px;
}
.tool-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 10px;
  border-radius: 6px;
  border: 1px solid #e5e6eb;
  background: #fff;
  font-size: 12px;
  color: #86909c;
  cursor: pointer;
  transition: all 0.15s;
}
.tool-chip:hover {
  background: #f2f3f5;
  border-color: #c9cdd4;
  color: #4e5969;
}
.tool-chip.active {
  background: #e8f1ff;
  border-color: #2468f2;
  color: #2468f2;
}
.tool-chip.send-btn.active {
  background: #2468f2;
  border-color: #2468f2;
  color: #fff;
}
.tool-chip.send-btn.active:hover {
  background: #1d5bd7;
  border-color: #1d5bd7;
}
.tool-plus {
  padding: 3px 6px;
  font-size: 16px;
  line-height: 1;
}

.tools-left :deep(.el-select) {
  --el-select-border-color-hover: transparent;
  --el-select-input-focus-border-color: transparent;
}
.tools-left :deep(.el-select .el-input__wrapper) {
  background: #f2f3f5;
  border-radius: 8px;
  box-shadow: none !important;
}
.tools-left :deep(.el-select .el-input__inner) {
  font-size: 12px;
  color: #4e5969;
}
.tools-right :deep(.el-select) {
  --el-select-border-color-hover: transparent;
  --el-select-input-focus-border-color: transparent;
}
.tools-right :deep(.el-select .el-input__wrapper) {
  background: #f2f3f5;
  border-radius: 8px;
  box-shadow: none !important;
}
.tools-right :deep(.el-select .el-input__inner) {
  font-size: 12px;
  color: #4e5969;
}
</style>

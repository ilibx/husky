<script setup lang="ts">
import { ref, onMounted, nextTick, computed } from 'vue'
import request from '@/api/request'

interface Message {
  role: 'user' | 'assistant'
  content: string
  sources?: { title: string; content: string; score: number }[]
  images?: string[]
  model?: string
  loading?: boolean
}

interface ModelOption {
  name: string
  provider: string
  enabled: boolean
  max_tokens: number
  temperature: number
}

const messages = ref<Message[]>([])
const inputText = ref('')
const sending = ref(false)
const chatBody = ref<HTMLElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const pendingImages = ref<string[]>([])
const selectedModel = ref('')
const deepThinking = ref(false)
const models = ref<ModelOption[]>([])
const modelLoading = ref(false)

function scrollToBottom() {
  nextTick(() => {
    if (chatBody.value) {
      chatBody.value.scrollTop = chatBody.value.scrollHeight
    }
  })
}

async function fetchModels() {
  modelLoading.value = true
  try {
    const res: any = await request.get('/system-config', { params: { page: 1, page_size: 200 } })
    const list: any[] = res.data?.data || []
    const result: ModelOption[] = []
    for (const item of list) {
      if (item.key && item.key.startsWith('model:')) {
        try {
          const val = JSON.parse(item.value)
          result.push({ name: item.key.slice(6), ...val })
        } catch { /* skip */ }
      }
    }
    models.value = result
    if (result.length > 0 && !selectedModel.value) {
      selectedModel.value = result[0].name
    }
  } catch {
    models.value = []
  } finally {
    modelLoading.value = false
  }
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

    const res: any = await request.post('/knowledge/ask', payload)
    const data = res.data
    messages.value[assistantIdx - 1].content = data?.answer || data?.content || '未获取到回答'
    messages.value[assistantIdx - 1].sources = data?.sources
    messages.value[assistantIdx - 1].model = data?.model
  } catch {
    messages.value[assistantIdx - 1].content = '查询失败，请稍后重试'
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
const selectedTool = ref('search')

function autoResize() {
  const el = textareaRef.value
  if (!el) return
  el.style.height = ''
  el.style.height = el.scrollHeight + 'px'
}

const canSend = computed(() => inputText.value.trim() !== '' || pendingImages.value.length > 0)

onMounted(fetchModels)
</script>

<template>
  <div class="chat-page">
    <div v-if="messages.length === 0" class="chat-empty">
      <div class="empty-icon">✨</div>
      <h2>有什么可以帮助你的？</h2>
      <p>基于文档库和知识库中的知识，向我提问</p>
    </div>

    <div v-else ref="chatBody" class="chat-body">
      <div v-for="(msg, idx) in messages" :key="idx" class="msg-row" :class="msg.role">
        <div class="msg-avatar">{{ msg.role === 'user' ? 'U' : 'A' }}</div>
        <div class="msg-main">
          <div class="msg-bubble">
            <div v-if="msg.images?.length" class="msg-imgs">
              <img v-for="(img, ii) in msg.images" :key="ii" :src="img" />
            </div>
            <div v-if="msg.loading" class="typing"><span/><span/><span/></div>
            <div v-else class="msg-text">{{ msg.content }}</div>
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
              style="width:100px"
            >
              <el-option label="搜索知识库" value="search" />
              <el-option label="联网搜索" value="web" />
            </el-select>
          </div>
          <div class="tools-right">
            <el-select
              v-model="selectedModel"
              placeholder="选择模型"
              size="small"
              :loading="modelLoading"
              style="width:140px"
              clearable
            >
              <el-option v-for="m in models" :key="m.name" :label="m.name" :value="m.name">
                <span>{{ m.name }}</span>
                <el-tag v-if="m.provider" size="small" type="info" style="margin-left:6px">{{ m.provider }}</el-tag>
              </el-option>
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

/* ---- Empty state ---- */
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

/* ---- Chat body ---- */
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
  white-space: pre-wrap;
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
  white-space: pre-wrap;
}
.typing {
  display: flex;
  gap: 5px;
  padding: 6px 0;
}
.typing span {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #c9cdd4;
  animation: blink 1.4s infinite both;
}
.typing span:nth-child(2) { animation-delay: 0.2s; }
.typing span:nth-child(3) { animation-delay: 0.4s; }
@keyframes blink {
  0%, 80%, 100% { opacity: 0.3; }
  40% { opacity: 1; }
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

/* ---- Footer ---- */
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

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Marked } from 'marked'
import { List, Link, EditPen, View } from '@element-plus/icons-vue'

const props = withDefaults(defineProps<{
  modelValue: string
  placeholder?: string
  minHeight?: string
}>(), {
  placeholder: '输入 Markdown 内容...',
  minHeight: '300px',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const marked = new Marked({ gfm: true, breaks: true })
const activeTab = ref<'edit' | 'preview'>('edit')

const renderedHTML = computed(() => {
  if (!props.modelValue) return ''
  try {
    const text = props.modelValue
    // Detect YAML frontmatter: first line is '---', then key:value lines, then '---'
    const lines = text.split('\n')
    if (lines.length > 1 && lines[0].trim() === '---') {
      let endIdx = -1
      for (let i = 1; i < lines.length; i++) {
        if (lines[i].trim() === '---') { endIdx = i; break }
      }
      if (endIdx > 0) {
        // Build YAML table from frontmatter lines
        let tableHtml = '<table class="yaml-table">'
        for (let i = 1; i < endIdx; i++) {
          const line = lines[i]
          const colonIdx = line.indexOf(':')
          if (colonIdx < 0) continue
          const key = line.slice(0, colonIdx).trim()
          const val = line.slice(colonIdx + 1).trim()
          tableHtml += `<tr><td class="yaml-key">${key}</td><td class="yaml-val">${val}</td></tr>`
        }
        tableHtml += '</table>'
        // Parse the rest as markdown
        const body = lines.slice(endIdx + 1).join('\n')
        const bodyHtml = marked.parse(body) as string
        return tableHtml + bodyHtml
      }
    }
    return marked.parse(text) as string
  } catch {
    return props.modelValue
  }
})

function onInput(e: Event) {
  const target = e.target as HTMLTextAreaElement
  emit('update:modelValue', target.value)
}

function insertMarkdown(before: string, after = '') {
  const textarea = document.querySelector('.md-editor-textarea') as HTMLTextAreaElement
  if (!textarea) return
  const start = textarea.selectionStart
  const end = textarea.selectionEnd
  const selected = props.modelValue.substring(start, end)
  const newText = props.modelValue.substring(0, start) + before + selected + after + props.modelValue.substring(end)
  emit('update:modelValue', newText)
  requestAnimationFrame(() => {
    textarea.focus()
    textarea.setSelectionRange(start + before.length, start + before.length + selected.length)
  })
}
</script>

<template>
  <div class="md-editor">
    <div class="md-editor-toolbar">
      <el-button-group>
        <el-button size="small" @click="insertMarkdown('**', '**')" title="加粗">
          <b>B</b>
        </el-button>
        <el-button size="small" @click="insertMarkdown('*', '*')" title="斜体">
          <i>I</i>
        </el-button>
        <el-button size="small" @click="insertMarkdown('# ', '')" title="标题1">
          <b>H1</b>
        </el-button>
        <el-button size="small" @click="insertMarkdown('## ', '')" title="标题2">
          <b>H2</b>
        </el-button>
        <el-button size="small" @click="insertMarkdown('### ', '')" title="标题3">
          <b>H3</b>
        </el-button>
        <el-button size="small" @click="insertMarkdown('- ', '')" title="无序列表">
          <el-icon><List /></el-icon>
        </el-button>
        <el-button size="small" @click="insertMarkdown('1. ', '')" title="有序列表">
          1.
        </el-button>
        <el-button size="small" @click="insertMarkdown('```\n', '\n```')" title="代码块">
          <span style="font-family:monospace;font-weight:700">&lt;/&gt;</span>
        </el-button>
        <el-button size="small" @click="insertMarkdown('[', '](url)')" title="链接">
          <el-icon><Link /></el-icon>
        </el-button>
        <el-button size="small" @click="insertMarkdown('> ', '')" title="引用">
          ❝
        </el-button>
      </el-button-group>
      <div class="md-editor-tabs">
        <el-button-group>
          <el-button
            :type="activeTab === 'edit' ? 'primary' : 'default'"
            size="small"
            @click="activeTab = 'edit'"
            title="编辑"
          >
            <el-icon><EditPen /></el-icon>
          </el-button>
          <el-button
            :type="activeTab === 'preview' ? 'primary' : 'default'"
            size="small"
            @click="activeTab = 'preview'"
            title="预览"
          >
            <el-icon><View /></el-icon>
          </el-button>
        </el-button-group>
      </div>
    </div>
    <div class="md-editor-body" :style="{ minHeight: props.minHeight }">
      <textarea
        v-show="activeTab === 'edit'"
        class="md-editor-textarea"
        :value="props.modelValue"
        @input="onInput"
        :placeholder="props.placeholder"
        :style="{ minHeight: props.minHeight }"
      />
      <div
        v-show="activeTab === 'preview'"
        class="md-editor-preview markdown-body"
        v-html="renderedHTML"
        :style="{ minHeight: props.minHeight }"
      />
    </div>
    <div class="md-editor-footer">
      <span class="md-editor-word-count">{{ props.modelValue.length }} 字</span>
    </div>
  </div>
</template>

<style scoped>
.md-editor {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
}
.md-editor-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 8px;
  background: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
  flex-wrap: wrap;
  gap: 4px;
}
.md-editor-tabs {
  display: flex;
  gap: 4px;
}
.md-editor-body {
  position: relative;
}
.md-editor-textarea {
  width: 100%;
  min-height: 300px;
  padding: 12px;
  border: none;
  outline: none;
  resize: vertical;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 14px;
  line-height: 1.6;
  background: #fff;
  color: #303133;
}
.md-editor-textarea:focus {
  background: #fafafa;
}
.md-editor-preview {
  padding: 12px;
  min-height: 300px;
  overflow-y: auto;
  background: #fff;
  line-height: 1.8;
}
.md-editor-footer {
  padding: 4px 12px;
  background: #f5f7fa;
  border-top: 1px solid #e4e7ed;
  text-align: right;
}
.md-editor-word-count {
  font-size: 12px;
  color: #909399;
}
.markdown-body {
  white-space: normal;
}
.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3) {
  margin: 16px 0 8px;
}
.markdown-body :deep(p) {
  margin: 8px 0;
}
.markdown-body :deep(code) {
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 90%;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
}
.markdown-body :deep(pre) {
  background: #f5f5f5;
  padding: 12px;
  border-radius: 4px;
  overflow-x: auto;
}
.markdown-body :deep(pre code) {
  background: none;
  padding: 0;
}
.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  padding-left: 20px;
}
.markdown-body :deep(blockquote) {
  border-left: 4px solid #ddd;
  padding-left: 12px;
  color: #666;
  margin: 8px 0;
}
.markdown-body :deep(table) {
  border-collapse: collapse;
  width: 100%;
}
.markdown-body :deep(th),
.markdown-body :deep(td) {
  border: 1px solid #ddd;
  padding: 6px 10px;
  text-align: left;
}
.markdown-body :deep(img) {
  max-width: 100%;
}
</style>

<style>
/* YAML frontmatter table in preview — unscoped because v-html lacks data attr */
.yaml-table {
  width: 100%;
  border-collapse: collapse;
  margin-bottom: 16px;
  background: #fafafa;
  border-radius: 6px;
  overflow: hidden;
}
.yaml-table tr {
  border-bottom: 1px solid #e5e6eb;
}
.yaml-table tr:last-child {
  border-bottom: none;
}
.yaml-table td {
  padding: 8px 12px;
  font-size: 13px;
  line-height: 1.5;
}
.yaml-key {
  width: 120px;
  font-weight: 600;
  color: #4e5969;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  vertical-align: middle;
  background: #f2f3f5;
}
.yaml-val {
  color: #303133;
}
</style>
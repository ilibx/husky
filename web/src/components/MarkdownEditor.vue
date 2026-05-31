<script setup lang="ts">
import { ref, computed } from 'vue'
import { Marked } from 'marked'

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
    return marked.parse(props.modelValue) as string
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
          <el-icon><Bold /></el-icon>
        </el-button>
        <el-button size="small" @click="insertMarkdown('*', '*')" title="斜体">
          <el-icon><Italic /></el-icon>
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
          <el-icon><OrderedList /></el-icon>
        </el-button>
        <el-button size="small" @click="insertMarkdown('```\n', '\n```')" title="代码块">
          <el-icon><Code /></el-icon>
        </el-button>
        <el-button size="small" @click="insertMarkdown('[', '](url)')" title="链接">
          <el-icon><Link /></el-icon>
        </el-button>
        <el-button size="small" @click="insertMarkdown('> ', '')" title="引用">
          <el-icon><Quote /></el-icon>
        </el-button>
      </el-button-group>
      <div class="md-editor-tabs">
        <el-button
          :type="activeTab === 'edit' ? 'primary' : 'default'"
          size="small"
          @click="activeTab = 'edit'"
        >编辑</el-button>
        <el-button
          :type="activeTab === 'preview' ? 'primary' : 'default'"
          size="small"
          @click="activeTab = 'preview'"
        >预览</el-button>
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
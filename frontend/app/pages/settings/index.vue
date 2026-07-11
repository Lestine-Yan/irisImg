<template>
  <div class="space-y-6">
    <!-- 顶部标题 -->
    <div>
      <h1 class="text-2xl font-bold text-gray-900">系统配置</h1>
      <p class="mt-1 text-sm text-gray-500">查看当前运行配置（只读）</p>
    </div>

    <!-- 固定提示：配置不可在线变更 -->
    <div class="flex items-start gap-3 rounded-2xl border border-amber-200 bg-amber-50 p-4">
      <svg class="mt-0.5 h-5 w-5 flex-shrink-0 text-amber-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="1.5"
          d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m2.25 0H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z"
        />
      </svg>
      <p class="text-sm text-amber-800">生产中无法变更配置，请修改 config 文件并重启以应用。</p>
    </div>

    <!-- 加载态 -->
    <div v-if="loading" class="rounded-2xl border border-dashed border-gray-200 bg-white/60 p-12 text-center">
      <p class="text-sm text-gray-400">加载中…</p>
    </div>

    <!-- 错误态 -->
    <div v-else-if="error" class="rounded-2xl border border-dashed border-rose-200 bg-rose-50 p-6 text-center">
      <p class="text-sm text-rose-600">{{ error }}</p>
      <button
        type="button"
        class="mt-3 rounded-xl border border-gray-200 px-3 py-1.5 text-sm text-gray-600 transition hover:bg-gray-50"
        @click="fetchConfig"
      >
        重试
      </button>
    </div>

    <!-- 数据态 -->
    <template v-else-if="config">
      <!-- https_only=false 警告 -->
      <div v-if="!config.apikey.https_only" class="flex items-start gap-3 rounded-2xl border border-rose-200 bg-rose-50 p-4">
        <svg class="mt-0.5 h-5 w-5 flex-shrink-0 text-rose-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.5"
            d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z"
          />
        </svg>
        <div class="text-sm text-rose-800">
          <p class="font-medium">https_only 当前为 false</p>
          <p class="mt-0.5">密钥相关敏感接口未强制 HTTPS，生产环境请在 config 中置 <code class="rounded bg-rose-100 px-1.5 py-0.5 font-mono text-xs">apikey.https_only: true</code>。</p>
        </div>
      </div>

      <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <!-- 服务 -->
        <SettingsConfigSection title="服务">
          <SettingsConfigItem label="监听地址">{{ config.server.host }}:{{ config.server.port }}</SettingsConfigItem>
        </SettingsConfigSection>

        <!-- 数据库 -->
        <SettingsConfigSection title="数据库">
          <SettingsConfigItem label="驱动">{{ config.database.driver }}</SettingsConfigItem>
          <SettingsConfigItem label="位置">{{ config.database.path }}</SettingsConfigItem>
        </SettingsConfigSection>

        <!-- APIKey -->
        <SettingsConfigSection title="APIKey">
          <SettingsConfigItem label="默认限速">{{ config.apikey.rate_limit_per_minute }} 次/分钟</SettingsConfigItem>
          <SettingsConfigItem label="HTTPS 校验">
            <span v-if="config.apikey.https_only" class="inline-flex items-center rounded-full bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-700">已启用</span>
            <span v-else class="inline-flex items-center rounded-full bg-rose-50 px-2 py-0.5 text-xs font-medium text-rose-700">未启用</span>
          </SettingsConfigItem>
        </SettingsConfigSection>

        <!-- 存储 -->
        <SettingsConfigSection title="存储">
          <SettingsConfigItem label="根目录">{{ config.storage.root_dir }}</SettingsConfigItem>
          <SettingsConfigItem label="公访问基址">
            <span v-if="config.storage.public_base_url">{{ config.storage.public_base_url }}</span>
            <span v-else class="text-gray-400">未设置（使用相对路径 /imgs/）</span>
          </SettingsConfigItem>
          <SettingsConfigItem label="上传上限">{{ config.storage.max_upload_size_mb }} MiB</SettingsConfigItem>
          <SettingsConfigItem label="允许类型">
            <span v-if="config.storage.allowed_mime_types.length" class="inline-flex flex-wrap justify-end gap-1.5">
              <span
                v-for="m in config.storage.allowed_mime_types"
                :key="m"
                class="inline-flex items-center rounded-full bg-iris-violet/10 px-2 py-0.5 text-xs font-medium text-iris-dark"
              >{{ m }}</span>
            </span>
            <span v-else class="text-gray-400">无</span>
          </SettingsConfigItem>
        </SettingsConfigSection>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import type { SystemConfig } from '~/composables/useSystemConfig'

definePageMeta({ middleware: 'auth' })

const { load } = useSystemConfig()

const config = ref<SystemConfig | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)

async function fetchConfig() {
  loading.value = true
  error.value = null
  try {
    config.value = await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : '加载配置失败'
    config.value = null
  } finally {
    loading.value = false
  }
}

onMounted(fetchConfig)
</script>

<template>
  <!-- Mobile: always-visible search (Products) + horizontal category chips -->
  <div class="lg:hidden min-w-0 space-y-3">
    <div v-if="showSearch" class="rounded-2xl border bg-card/95 p-3">
      <label class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        {{ t('products.searchLabel') }}
      </label>
      <div class="mt-2 flex items-center gap-2">
        <Input
          :model-value="searchQuery"
          @update:model-value="$emit('update:searchQuery', String($event))"
          type="text"
          class="min-w-0 flex-1"
          :placeholder="t('products.searchPlaceholder')"
        />
        <Button
          v-if="searchQuery"
          variant="secondary"
          size="sm"
          type="button"
          class="shrink-0 whitespace-nowrap rounded-xl text-xs"
          @click="$emit('clearSearch')"
        >
          {{ t('common.cancel') }}
        </Button>
      </div>
    </div>

    <div
      class="flex gap-1.5 overflow-x-auto overscroll-x-contain pb-0.5 [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
      role="listbox"
      :aria-label="t('products.categories')"
    >
      <button
        type="button"
        role="option"
        :aria-selected="selectedCategory === null"
        class="flex-none whitespace-nowrap rounded-full px-3.5 py-2 text-sm font-semibold transition-colors active:scale-[0.97]"
        :class="selectedCategory === null
          ? 'bg-primary text-primary-foreground'
          : 'bg-secondary text-muted-foreground hover:text-foreground'"
        @click="$emit('selectCategory', null)"
      >
        {{ t('products.allCategories') }}
      </button>
      <button
        v-for="group in categories"
        :key="group.id"
        type="button"
        role="option"
        :aria-selected="selectedCategory === group.id"
        class="flex max-w-[70vw] flex-none items-center gap-2 whitespace-nowrap rounded-full px-3.5 py-2 text-sm font-semibold transition-colors active:scale-[0.97]"
        :class="selectedCategory === group.id
          ? 'bg-primary text-primary-foreground'
          : 'bg-secondary text-muted-foreground hover:text-foreground'"
        @click="$emit('selectCategory', group.id)"
      >
        <img
          v-if="group.icon"
          :src="getImageUrl(group.icon)"
          :alt="getLocalizedText(group.name)"
          loading="lazy"
          class="h-5 w-5 flex-none rounded object-cover"
        />
        <span class="min-w-0 truncate">{{ getLocalizedText(group.name) }}</span>
      </button>
    </div>
  </div>

  <!-- Desktop Sidebar -->
  <aside class="hidden lg:block flex-shrink-0" :class="compact ? 'lg:w-60' : 'lg:w-64'">
    <div class="bg-card backdrop-blur-xl border rounded-2xl sticky top-24" :class="compact ? 'p-5' : 'p-6'">
      <!-- Search (desktop, only for Products page) -->
      <div v-if="showSearch" :class="compact ? 'mb-4' : 'mb-6'">
        <label class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          {{ t('products.searchLabel') }}
        </label>
        <div class="mt-3 flex items-center gap-2">
          <Input :model-value="searchQuery" @update:model-value="$emit('update:searchQuery', String($event))" type="text"
            class="min-w-0 flex-1"
            :placeholder="t('products.searchPlaceholder')" />
          <Button v-if="searchQuery" variant="secondary" size="sm" type="button" class="shrink-0 whitespace-nowrap rounded-xl text-xs" @click="$emit('clearSearch')">
            {{ t('common.cancel') }}
          </Button>
        </div>
      </div>

      <h2 :class="compact ? 'text-base font-bold mb-4' : 'text-lg font-bold mb-6'"
        class="text-foreground flex items-center gap-2">
        <span class="w-1 h-5 bg-primary rounded-full"></span>
        {{ t('products.categories') }}
      </h2>

      <ul :class="compact ? 'space-y-1.5' : 'space-y-2'">
        <li>
          <button @click="$emit('selectCategory', null)"
            class="w-full text-left rounded-xl transition-all duration-300 border"
            :class="[
              compact ? 'px-3 py-2.5 text-sm' : 'px-4 py-3',
              selectedCategory === null
                ? 'bg-primary text-primary-foreground border border-transparent'
                : 'border-transparent text-muted-foreground hover:text-foreground hover:bg-secondary'
            ]">
            {{ t('products.allCategories') }}
          </button>
        </li>
        <li v-for="group in categories" :key="group.id">
          <div :class="compact ? 'space-y-1.5' : 'space-y-2'">
            <div class="flex items-stretch" :class="compact ? 'gap-1.5' : 'gap-2'">
              <button @click="$emit('selectCategory', group.id)"
                class="flex-1 min-w-0 text-left rounded-xl transition-all duration-300 border flex items-center gap-2"
                :class="[
                  compact ? 'px-3 py-2.5 text-sm' : 'px-4 py-3',
                  selectedCategory === group.id
                    ? 'bg-primary text-primary-foreground border border-transparent'
                    : 'border-transparent text-muted-foreground hover:text-foreground hover:bg-secondary'
                ]">
                <img v-if="group.icon" :src="getImageUrl(group.icon)"
                  :alt="getLocalizedText(group.name)"
                  loading="lazy" :class="compact ? 'h-4 w-4' : 'h-5 w-5'"
                  class="rounded object-cover" />
                <span class="truncate">{{ getLocalizedText(group.name) }}</span>
              </button>
              <button
                v-if="group.children.length > 0"
                type="button"
                class="shrink-0 self-center rounded-full border flex items-center justify-center shadow-sm transition-all duration-300 hover:-translate-y-0.5 hover:shadow-md"
                :class="[
                  compact ? 'h-9 w-9' : 'h-10 w-10',
                  expandedParentIds.includes(group.id) ? 'bg-primary text-primary-foreground border-transparent' : 'bg-card/70 text-muted-foreground hover:text-foreground'
                ]"
                @click.stop="$emit('toggleParent', group.id)"
              >
                <ChevronRight :class="[
                    compact ? 'w-3.5 h-3.5' : 'w-4 h-4',
                    expandedParentIds.includes(group.id) ? 'rotate-90' : ''
                  ]"
                  class="transition-transform duration-200" />
              </button>
            </div>
            <ul v-if="group.children.length > 0 && expandedParentIds.includes(group.id)"
              :class="compact ? 'space-y-1.5 pl-3' : 'space-y-2 pl-4'">
              <li v-for="child in group.children" :key="child.id">
                <button @click="$emit('selectCategory', child.id)"
                  class="w-full text-left rounded-xl transition-all duration-300 border flex items-center gap-2"
                  :class="[
                    compact ? 'px-3 py-2.5 text-sm' : 'px-4 py-3',
                    selectedCategory === child.id
                      ? 'bg-primary text-primary-foreground border border-transparent'
                      : 'border-transparent text-muted-foreground hover:text-foreground hover:bg-secondary'
                  ]">
                  <img v-if="child.icon" :src="getImageUrl(child.icon)"
                    :alt="getLocalizedText(child.name)"
                    loading="lazy" :class="compact ? 'h-4 w-4' : 'h-5 w-5'"
                    class="rounded object-cover" />
                  <span class="truncate">{{ getLocalizedText(child.name) }}</span>
                </button>
              </li>
            </ul>
          </div>
        </li>
      </ul>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { ChevronRight } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { getImageUrl } from '../utils/image'
import { useLocalized } from '../composables/useProduct'
import type { CategoryGroup } from '../utils/category'

const { t } = useI18n()
const { getLocalizedText } = useLocalized()

defineProps<{
  categories: CategoryGroup[]
  selectedCategory: number | null
  expandedParentIds: number[]
  /** Kept for parent binding compatibility; mobile drawer removed. */
  showDrawer: boolean
  compact?: boolean
  showSearch?: boolean
  searchQuery?: string
}>()

defineEmits<{
  'selectCategory': [categoryId: number | null, closeDrawer?: boolean]
  'toggleParent': [categoryId: number]
  'update:showDrawer': [value: boolean]
  'update:searchQuery': [value: string]
  'clearSearch': []
}>()
</script>

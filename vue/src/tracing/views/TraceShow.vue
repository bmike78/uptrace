<template>
  <XPlaceholder>
    <template v-if="trace.error" #placeholder>
      <TraceError v-if="trace.error" :error="trace.error" />
    </template>

    <template v-else-if="!trace.root" #placeholder>
      <v-container :fluid="$vuetify.breakpoint.mdAndDown">
        <v-skeleton-loader type="article, image, table" />
      </v-container>
    </template>

    <PageToolbar :loading="trace.loading" :fluid="$vuetify.breakpoint.mdAndDown">
      <v-breadcrumbs :items="meta.breadcrumbs" divider=">" large>
        <template #item="{ item }">
          <v-breadcrumbs-item :to="item.to" :exact="item.exact">
            {{ item.text }}
          </v-breadcrumbs-item>
        </template>
      </v-breadcrumbs>

      <v-spacer />

      <FixedDateRangePicker
        v-if="trace.root"
        :date-range="dateRange"
        :around="trace.root.time"
        show-reload
      />
    </PageToolbar>

    <v-container :fluid="$vuetify.breakpoint.mdAndDown" class="py-4">
      <v-row v-if="trace.root" class="px-4 text-body-2">
        <v-col class="word-break-all">
          {{ eventOrSpanName(trace.root.name) }}
        </v-col>
      </v-row>

      <v-row v-if="trace.root" align="end" class="px-4 text-subtitle-2 text-center">
        <v-col v-if="trace.root.kind" cols="auto">
          <div class="grey--text font-weight-regular">Kind</div>
          <div>{{ trace.root.kind }}</div>
        </v-col>

        <v-col v-if="trace.root.statusCode" cols="auto">
          <div class="grey--text font-weight-regular">Status</div>
          <div :class="{ 'error--text': trace.root.statusCode === 'error' }">
            {{ trace.root.statusCode }}
          </div>
        </v-col>

        <v-col cols="auto">
          <div class="grey--text font-weight-regular">Time</div>
          <XDate :date="trace.root.time" format="full" />
        </v-col>

        <v-col cols="auto">
          <div class="grey--text font-weight-regular">Duration</div>
          <XDuration :duration="trace.root.duration" fixed />
        </v-col>

        <v-col cols="auto">
          <v-btn v-if="groupRoute" depressed small :to="groupRoute" exact>View group</v-btn>
          <v-btn v-if="exploreTraceRoute" depressed small :to="exploreTraceRoute" exact class="ml-2"
            >Explore trace</v-btn
          >
        </v-col>
      </v-row>

      <v-row v-if="trace.root && trace.root.groupId">
        <v-col>
          <v-card outlined rounded="lg">
            <v-card-text>
              <LoadPctileChart :axios-params="axiosParams" />
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <SystemBarChart :systems="trace.coloredSystems" />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <TraceTabs :date-range="dateRange" :trace="trace" />
        </v-col>
      </v-row>
    </v-container>
  </XPlaceholder>
</template>

<script lang="ts">
import { defineComponent, computed, watch, proxyRefs } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { useDateRange, UseDateRange } from '@/use/date-range'
import { useTrace, UseTrace } from '@/tracing/use-trace'
import { createUqlEditor } from '@/use/uql'

// Components
import FixedDateRangePicker from '@/components/date/FixedDateRangePicker.vue'
import LoadPctileChart from '@/components/LoadPctileChart.vue'
import SystemBarChart from '@/components/SystemBarChart.vue'
import TraceTabs from '@/tracing/TraceTabs.vue'
import TraceError from '@/tracing/TraceError.vue'

// Utilities
import { xkey } from '@/models/otelattr'
import { eventOrSpanName } from '@/models/span'

export default defineComponent({
  name: 'TraceShow',
  components: {
    FixedDateRangePicker,
    LoadPctileChart,
    SystemBarChart,
    TraceTabs,
    TraceError,
  },

  setup() {
    useTitle('View trace')
    const dateRange = useDateRange()
    const trace = useTrace()

    const axiosParams = computed(() => {
      if (!trace.root) {
        return
      }
      return {
        ...dateRange.axiosParams(),
        system: trace.root.system,
        group_id: trace.root.groupId,
      }
    })

    const groupRoute = computed(() => {
      if (!trace.root) {
        return
      }

      return {
        name: 'SpanList',
        query: {
          ...dateRange.queryParams(),
          system: trace.root.system,
          query: createUqlEditor()
            .exploreAttr(xkey.spanGroupId)
            .where(xkey.spanGroupId, '=', trace.root.groupId)
            .toString(),
        },
      }
    })

    const exploreTraceRoute = computed(() => {
      if (!trace.root) {
        return
      }

      return {
        name: 'SpanGroupList',
        query: {
          ...dateRange.queryParams(),
          system: xkey.allSystem,
          query: [
            `where ${xkey.spanTraceId} = ${trace.root.traceId}`,
            `group by ${xkey.spanGroupId}`,
            xkey.spanCount,
            xkey.spanErrorCount,
            `{p50,p90,p99,sum}(${xkey.spanDuration})`,
          ].join(' | '),
          plot: null,
        },
      }
    })

    watch(
      () => trace.root,
      (span) => {
        if (span) {
          useTitle(span.name)
        }
      },
    )

    return {
      dateRange,
      trace,
      meta: useMeta(dateRange, trace),

      axiosParams,
      groupRoute,
      exploreTraceRoute,
      eventOrSpanName,
    }
  },
})

function useMeta(dateRange: UseDateRange, trace: UseTrace) {
  const breadcrumbs = computed(() => {
    const root = trace.root
    const bs: any[] = []

    if (!root) {
      return bs
    }

    if (root.system) {
      bs.push({
        text: root.system,
        to: {
          name: 'SpanGroupList',
          query: {
            ...dateRange.queryParams(),
            system: root.system,
          },
        },
        exact: true,
      })
    }

    if (root.system && root.groupId) {
      bs.push({
        text: eventOrSpanName(root, 60),
        to: {
          name: 'SpanList',
          query: {
            ...dateRange.queryParams(),
            system: root.system,
            query: createUqlEditor()
              .exploreAttr(xkey.spanGroupId)
              .where(xkey.spanGroupId, '=', root.groupId)
              .toString(),
          },
        },
        exact: true,
      })
    }

    bs.push({
      text: trace.id,
    })

    return bs
  })

  return proxyRefs({ breadcrumbs })
}
</script>

<style lang="scss" scoped></style>

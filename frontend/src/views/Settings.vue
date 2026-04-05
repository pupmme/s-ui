<template>
  <v-card :loading="loading">
    <v-tabs
    v-model="tab"
    color="primary"
    align-tabs="center"
    show-arrows
  >
    <v-tab value="t1">{{ $t('setting.interface') }}</v-tab>
              </v-tabs>
  <v-card-text>
    <v-row align="center" justify="center" style="margin-bottom: 10px;">
      <v-col cols="auto">
        <v-btn color="primary" @click="save" :loading="loading" :disabled="!stateChange">
          {{ $t('actions.save') }}
        </v-btn>
      </v-col>
      <v-col cols="auto">
        <v-btn variant="outlined" color="warning" @click="restartApp" :loading="loading" :disabled="stateChange">
          {{ $t('actions.restartApp') }}
        </v-btn>
      </v-col>
    </v-row>
    <v-window v-model="tab">
      <v-window-item value="t1">
        <v-row>
          <v-col cols="12" sm="6" md="4">
            <v-text-field v-model="settings.webListen" :label="$t('setting.addr')" hide-details></v-text-field>
          </v-col>
          <v-col cols="12" sm="6" md="4">
            <v-text-field v-model.number="webPort" min="1" type="number" :label="$t('setting.port')" hide-details></v-text-field>
          </v-col>
          <v-col cols="12" sm="6" md="4">
            <v-text-field v-model="settings.webPath" :label="$t('setting.webPath')" hide-details></v-text-field>
          </v-col>
          <v-col cols="12" sm="6" md="4">
            <v-text-field v-model="settings.webDomain" :label="$t('setting.domain')" hide-details></v-text-field>
          </v-col>
          <v-col cols="12" sm="6" md="4">
            <v-text-field v-model="settings.webKeyFile" :label="$t('setting.sslKey')" hide-details></v-text-field>
          </v-col>
          <v-col cols="12" sm="6" md="4">
            <v-text-field v-model="settings.webCertFile" :label="$t('setting.sslCert')" hide-details></v-text-field>
          </v-col>
          <v-col cols="12" sm="6" md="4">
            <v-text-field v-model="settings.webURI" :label="$t('setting.webUri')" hide-details></v-text-field>
          </v-col>
          <v-col cols="12" sm="6" md="4">
            <v-text-field
              type="number"
              v-model.number="sessionMaxAge"
              min="0"
              :label="$t('setting.sessionAge')"
              :suffix="$t('date.m')"
              hide-details
              ></v-text-field>
          </v-col>
          <v-col cols="12" sm="6" md="4">
            <v-text-field
              type="number"
              v-model.number="trafficAge"
              min="0"
              :label="$t('setting.trafficAge')"
              :suffix="$t('date.d')"
              hide-details
              ></v-text-field>
          </v-col>
          <v-col cols="12" sm="6" md="4">
            <v-text-field v-model="settings.timeLocation" :label="$t('setting.timeLoc')" hide-details></v-text-field>
          </v-col>
        </v-row>

        <v-divider class="my-4" />

        <v-row align="center">
          <v-col cols="12" sm="6" md="4">
            <v-switch
              v-model="nodeMode"
              color="primary"
              :label="$t('setting.nodeMode')"
              hide-details
            ></v-switch>
          </v-col>
          <v-col cols="12" sm="6" md="4">
            <v-chip v-if="nodeMode" color="success" variant="flat">Node 模式</v-chip>
            <v-chip v-else color="info" variant="flat">Local 模式</v-chip>
          </v-col>
        </v-row>

        <v-expand-transition>
          <div v-if="nodeMode">
            <v-row>
              <v-col cols="12" sm="6" md="4">
                <v-text-field
                  v-model="settings.xboardApiHost"
                  :label="$t('setting.xboardApiHost')"
                  placeholder="https://s.pupm.us"
                  hide-details
                ></v-text-field>
              </v-col>
              <v-col cols="12" sm="6" md="4">
                <v-text-field
                  v-model="settings.xboardApiKey"
                  :label="$t('setting.xboardApiKey')"
                  placeholder="API Key"
                  hide-details
                ></v-text-field>
              </v-col>
              <v-col cols="12" sm="6" md="4">
                <v-text-field
                  v-model="settings.nodeId"
                  :label="$t('setting.nodeId')"
                  placeholder="1"
                  type="number"
                  hide-details
                ></v-text-field>
              </v-col>
              <v-col cols="12" sm="6" md="4">
                <v-select
                  v-model="settings.nodeType"
                  :label="$t('setting.nodeType')"
                  :items="['sing-box', 'xray', 'v2ray', 'shadowsocks', 'custom']"
                  hide-details
                ></v-select>
              </v-col>
            </v-row>
          </div>
        </v-expand-transition>
      </v-window-item>

                      </v-window>
  </v-card-text>
</v-card>
</template>

<script lang="ts" setup>
import { i18n } from '@/locales'
import { Ref, computed, inject, onMounted, ref } from 'vue'
import HttpUtils from '@/plugins/httputil'
import { FindDiff } from '@/plugins/utils'
import SubJsonExtVue from '@/components/SubJsonExt.vue'
import SubClashExtVue from '@/components/SubClashExt.vue'
import { push } from 'notivue'
const tab = ref("t1")
const loading:Ref = inject('loading')?? ref(false)
const oldSettings = ref({})

const settings = ref({
	webListen: "",
	webDomain: "",
	webPort: "2095",
	webCertFile: "",
	webKeyFile: "",
  webPath: "/app/",
  webURI: "",
	sessionMaxAge: "0",
  trafficAge: "30",
	timeLocation: "Asia/Tehran",
  subListen: "",
	subPort: "2096",
	subPath: "/sub/",
	subDomain: "",
	subCertFile: "",
	subKeyFile: "",
	subUpdates: "12",
	subEncode: "true",
	subShowInfo: "false",
	subURI: "",
  subJsonExt: "",
  subClashExt: "",
  nodeMode: "false",
  xboardApiHost: "",
  xboardApiKey: "",
  nodeId: 0,
  nodeType: "sing-box",
})

onMounted(async () => {
  loading.value = true
  await loadData()
  loading.value = false
})

const loadData = async () => {
  loading.value = true
  const [settingsMsg, nodeMsg] = await Promise.all([
    HttpUtils.get('api/settings'),
    HttpUtils.get('api/getNodeMode'),
  ])
  loading.value = false
  if (settingsMsg.success) {
    setData(settingsMsg.obj)
  }
  if (nodeMsg.success) {
    settings.value.nodeMode = nodeMsg.obj.nodeMode
    settings.value.xboardApiHost = nodeMsg.obj.xboardApiHost || ''
    settings.value.xboardApiKey = nodeMsg.obj.xboardApiKey || ''
    settings.value.nodeId = nodeMsg.obj.nodeId || 0
    settings.value.nodeType = nodeMsg.obj.nodeType || 'sing-box'
  }
  oldSettings.value = { ...settings.value }
}

const setData = (data: any) => {
  settings.value = { ...settings.value, ...data }
}

const save = async () => {
  loading.value = true
  // Separate node/config fields from panel settings
  const nodeFields = ['nodeMode', 'xboardApiHost', 'xboardApiKey', 'nodeId', 'nodeType']
  const nodeData: any = {}
  const panelData: any = {}
  for (const [k, v] of Object.entries(settings.value)) {
    if (nodeFields.includes(k)) nodeData[k] = v
    else panelData[k] = v
  }
  // Save node config first
  if (Object.keys(nodeData).length > 0) {
    await HttpUtils.post('api/setNodeMode', nodeData)
  }
  // Save panel settings
  const msg = await HttpUtils.post('api/save', { object: 'settings', action: 'set', data: JSON.stringify(panelData) })
  if (msg.success) {
    push.success({
      title: i18n.global.t('success'),
      duration: 5000,
      message: i18n.global.t('actions.set') + " " + i18n.global.t('pages.settings')
    })
    if (msg.obj && msg.obj.settings) setData(msg.obj.settings)
  }
  loading.value = false
}

const sleep = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

const restartApp = async () => {
  loading.value = true
  const msg = await HttpUtils.post('api/restartApp',{})
  if (msg.success) {
    let url = settings.value.webURI
    if (url !== "") {
      const isTLS = settings.value.webCertFile !== "" || settings.value.webKeyFile !== ""
      url = buildURL(settings.value.webDomain,settings.value.webPort.toString(),isTLS, settings.value.webPath)
    }
    await sleep(3000)
    window.location.replace(url)
  }
  loading.value = false
}

const buildURL = (host: string, port: string, isTLS: boolean, path: string) => {
  if (!host || host.length == 0) host = window.location.hostname
  if (!port || port.length == 0) port = window.location.port

  const protocol = isTLS ? "https:" : "http:"

  if (port === "" || (isTLS && port === "443") || (!isTLS && port === "80")) {
      port = ""
  } else {
      port = `:${port}`
  }

  return `${protocol}//${host}${port}${path}settings`
}

const subEncode = computed({
  get: () => { return settings.value.subEncode == "true" },
  set: (v:boolean) => { settings.value.subEncode = v ? "true" : "false" }
})

const subShowInfo = computed({
  get: () => { return settings.value.subShowInfo == "true" },
  set: (v:boolean) => { settings.value.subShowInfo = v ? "true" : "false" }
})

const webPort = computed({
  get: () => { return settings.value.webPort.length>0 ? parseInt(settings.value.webPort) : 2095 },
  set: (v:number) => { settings.value.webPort = v>0 ? v.toString() : "2095" }
})

const sessionMaxAge = computed({
  get: () => { return settings.value.sessionMaxAge.length>0 ? parseInt(settings.value.sessionMaxAge) : 0 },
  set: (v:number) => { settings.value.sessionMaxAge = v>0 ? v.toString() : "0" }
})

const trafficAge = computed({
  get: () => { return settings.value.trafficAge.length>0 ? parseInt(settings.value.trafficAge) : 0 },
  set: (v:number) => { settings.value.trafficAge = v>0 ? v.toString() : "0" }
})

const subPort = computed({
  get: () => { return settings.value.subPort.length>0 ? parseInt(settings.value.subPort) : 2096 },
  set: (v:number) => { settings.value.subPort = v>0 ? v.toString() : "2096" }
})

const subUpdates = computed({
  get: () => { return settings.value.subUpdates.length>0 ? parseInt(settings.value.subUpdates) : 12 },
  set: (v:number) => { settings.value.subUpdates = v>0 ? v.toString() : "12" }
})

const nodeMode = computed({
  get: () => !!settings.value.nodeMode,
  set: (v:boolean) => { settings.value.nodeMode = v }
})

const stateChange = computed(() => {
  return !FindDiff.deepCompare(settings.value,oldSettings.value)
})
</script>

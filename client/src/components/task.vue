<template>
  <v-container v-if="loading" fluid fill-height>
    <v-layout align-center justify-center>
      <fingerprint-spinner :animation-duration="2000" :size="200" :color="'#9FC657'"></fingerprint-spinner>
    </v-layout>
  </v-container>
  <v-container v-else class="pa-0" fluid fill-height>
    <v-layout column fill-height>
      <v-flex>
        <v-breadcrumbs v-if="project.id !== task.id">
          <v-breadcrumbs-item
            class="active"
            v-for="ancestor in ancestors"
            :key="ancestor.id"
            :disabled="false"
            :href="'/#/app/region/' + routeParams.region + '/shard/' + routeParams.shard + '/account/' + routeParams.account + '/project/' + project.id + '/task/' + ancestor.id"
          >
            {{ ancestor.name }}
          </v-breadcrumbs-item>
          <v-breadcrumbs-item> <!--empty breadcrumb item as the last item is always disaabled, not suer if this is a bug or a feature but it's not helpful here -->
          </v-breadcrumbs-item>
        </v-breadcrumbs>
        <v-card>
          <v-card-title class="py-1"><h3>{{task.name}}</h3></v-card-title>
          <v-card-text class="py-1">{{task.description}}</v-card-text>
          <v-card-text class="py-1">
            <div v-if="task.isAbstract">Parallel: <h3>{{task.isParallel}}</h3> </div>
            <div v-if="task.isAbstract">Min: <h3>{{ printDuration(task.minimumRemainingTime, false, project.hoursPerDay, project.daysPerWeek)}}</h3> </div>
            <div>Tot: <h3>{{printDuration(task.totalRemainingTime, false, project.hoursPerDay, project.daysPerWeek)}}</h3> </div>
            <div>Log: <h3>{{printDuration(task.totalLoggedTime, false, project.hoursPerDay, project.daysPerWeek)}}</h3> </div>
            <div v-if="task.isAbstract">Children: <h3>{{task.childCount}}</h3> </div>
            <div v-if="task.isAbstract">Descendants: <h3>{{task.descendantCount}}</h3> </div>
          </v-card-text>
        </v-card>
      </v-flex>
      <v-data-table
        v-if="task.isAbstract"
        style="width: 100%!important; height: 100%!important"
        :headers="childrenHeaders"
        :items="children"
        :loading="loading"
        hide-actions
        class="elevation-1"
      >
        <template slot="items" slot-scope="children">
          <tr @click="goToTask(children.item)" style="cursor: pointer">
            <td class="text-xs-left">{{ children.item.name }}</td>
            <td class="text-xs-left hidden-sm-and-down">{{ children.item.description? children.item.description: 'none' }}</td>
            <td class="text-xs-left" style="width: 120px;">{{ printDuration(children.item.minimumRemainingTime, false, project.hoursPerDay, project.daysPerWeek) }}</td>
            <td class="text-xs-left" style="width: 120px;">{{ printDuration(children.item.totalRemainingTime, false, project.hoursPerDay, project.daysPerWeek) }}</td>
            <td class="text-xs-left" style="width: 120px;">{{ printDuration(children.item.totalLoggedTime, false, project.hoursPerDay, project.daysPerWeek) }}</td>
          </tr>
        </template>
        <template slot="no-data">
          <v-alert v-if="!loading" :value="true" color="info" icon="error">
            No Child Tasks Yet <v-btn v-on:click="toggleCreateTaskForm" color="primary">create</v-btn>
          </v-alert>
          <v-alert v-if="loading" :value="true" color="info" icon="error">
            Loading Data
          </v-alert>
        </template>
      </v-data-table>
      <v-data-table
        v-else
        style="width: 100%!important; height: 100%!important"
        :headers="timeLogHeaders"
        :items="timeLogs"
        :loading="loading"
        hide-actions
        class="elevation-1"
      >
        <template slot="items" slot-scope="timeLogs">
          <tr>
            <td class="text-xs-left">{{ new Date(timeLogs.item.loggedOn).toLocaleDateString() }}</td>
            <td class="text-xs-left">{{ printDuration(timeLogs.item.duration, false, project.hoursPerDay, project.daysPerWeek) }}</td>
            <td class="text-xs-left">{{ timeLogs.item.note }}</td>
          </tr>
        </template>
        <template slot="no-data">
          <v-alert v-if="!loading" :value="true" color="info" icon="error">
            No Time Logs Yet <v-btn v-on:click="toggleCreateTimeLogForm" color="primary">create</v-btn>
          </v-alert>
          <v-alert v-if="loading" :value="true" color="info" icon="error">
            Loading Data
          </v-alert>
        </template>
      </v-data-table>
      <v-btn v-if="timeLogs.length > 0" v-on:click="toggleCreateTimeLogForm" fixed right bottom color="primary" fab>
        <v-icon>add</v-icon>
      </v-btn>
      <v-btn v-if="children.length > 0" v-on:click="toggleCreateTaskForm" fixed right bottom color="primary" fab>
        <v-icon>add</v-icon>
      </v-btn>
      <v-dialog
        v-model="createTaskDialog"
        fullscreen
        hide-overlay
        transition="dialog-bottom-transition"
        scrollable
      >
        <v-card tile>
          <v-toolbar card dark color="primary">
            <v-toolbar-title>Create Task</v-toolbar-title>
            <v-spacer></v-spacer>
            <v-btn icon dark @click.native="toggleCreateTaskForm">
              <v-icon>close</v-icon>
            </v-btn>
          </v-toolbar>
          <v-card-text>
            <v-form ref="form" @keyup.native.enter="createTask">
              <v-switch
                :label="`Is Abstract`"
                v-model="createTaskIsAbstract"></v-switch>
              <v-text-field v-model="createTaskName" name="taskName" label="Name" type="text"></v-text-field>
              <v-text-field v-model="createTaskDescription" name="taskDescription" label="Description" type="text"></v-text-field>
              <v-text-field v-if="!createTaskIsAbstract" v-model="createTaskRemainingTime" name="taskRemainingTime" label="Remaining Time" type="number"></v-text-field>
              <v-switch
                v-if="createTaskIsAbstract"
                :label="`Is Parallel`"
                v-model="createTaskIsParallel"></v-switch>
              <v-btn
                :loading="creatingTask"
                :disabled="creatingTask"
                color="secondary"
                @click="createTask"
              >
                Create
              </v-btn>
            </v-form>
          </v-card-text>
        </v-card>
      </v-dialog>
      <v-dialog
        v-model="createTimeLogDialog"
        fullscreen
        hide-overlay
        transition="dialog-bottom-transition"
        scrollable
      >
        <v-card tile>
          <v-toolbar card dark color="primary">
            <v-toolbar-title>Create Time Log</v-toolbar-title>
            <v-spacer></v-spacer>
            <v-btn icon dark @click.native="toggleCreateTimeLogForm">
              <v-icon>close</v-icon>
            </v-btn>
          </v-toolbar>
          <v-card-text>
            <v-form ref="form" @keyup.native.enter="createTimeLog">
              <v-text-field v-model="createTimeLogDuration" name="timeLogDuration" label="Duration" type="number"></v-text-field>
              <v-text-field v-model="createTimeLogRemainingTime" name="timeLogRemainingTime" label="Set Remaining Time" type="number"></v-text-field>
              <v-text-field v-model="createTimeLogNote" name="timeLogNote" label="Note" type="text"></v-text-field>
              <v-btn
                :loading="creatingTimeLog"
                :disabled="creatingTimeLog || createTimeLogDuration === null || createTimeLogDuration < 1"
                color="secondary"
                @click="createTimeLog"
              >
                Create
              </v-btn>
            </v-form>
          </v-card-text>
        </v-card>
      </v-dialog>
    </v-layout>
  </v-container>
</template>

<script>
  import api from '@/api'
  import router from '@/router'
  import {printDuration} from '@/helper'
  import {FingerprintSpinner} from 'epic-spinners'
  export default {
    name: 'task',
    components: {FingerprintSpinner},
    data () {
      let routeParams = router.currentRoute.params
      return {
        routeParams,
        project: null,
        childrenHeaders: [
          {text: 'Name', sortable: false, align: 'left', value: 'name'},
          {text: 'Description', class: 'hidden-sm-and-down', sortable: false, align: 'left', value: 'description'},
          {text: 'Min.', sortable: false, align: 'left', value: 'minimumRemainingTime'},
          {text: 'Tot.', sortable: false, align: 'left', value: 'totalRemainingTime'},
          {text: 'Log.', sortable: false, align: 'left', value: 'totalLoggedTime'}
        ],
        timeLogHeaders: [
          {text: 'Logged On', sortable: false, align: 'left', value: 'loggedOn'},
          {text: 'Duration', sortable: false, align: 'left', value: 'duration'},
          {text: 'Note', sortable: false, align: 'left', value: 'note'}
        ],
        pagination: {
          descending: false
        },
        loading: true,
        createTaskDialog: false,
        createTimeLogDialog: false,
        createTaskIsAbstract: true,
        createTaskName: '',
        createTaskDescription: null,
        createTaskRemainingTime: null,
        createTaskIsParallel: true,
        creatingTask: false,
        creatingTimeLog: false,
        createTimeLogDuration: null,
        createTimeLogRemainingTime: null,
        createTimeLogNote: null,
        children: [],
        timeLogs: [],
        moreTimeLogs: false,
        moreChildren: false,
        moreAncestors: false,
        task: null,
        ancestors: []
      }
    },
    mounted () {
      let htmlEl = document.querySelector('html')
      let self = this
      this.pageScrollListener = () => {
        if (self.moreChildren && htmlEl.clientHeight + htmlEl.scrollTop >= htmlEl.scrollHeight) {
          self.loadMore()
        }
      }
      document.addEventListener('scroll', this.pageScrollListener)
      this.init()
    },
    beforeDestroy () {
      document.removeEventListener('scroll', this.pageScrollListener)
    },
    watch: {
      $route () {
        this.init()
      }
    },
    methods: {
      printDuration,
      init () {
        let params = router.currentRoute.params
        let mapi = api.newMGetApi(params.region)
        mapi.v1.project.get(params.region, params.shard, params.account, params.project).then((project) => {
          this.project = project
          if (params.project === params.task) {
            this.task = project
            this.task.isAbstract = true
          }
        })
        if (params.project !== params.task) {
          mapi.v1.task.get(params.region, params.shard, params.account, params.project, params.task).then((task) => {
            this.task = task
          })
        }
        mapi.v1.task.getAncestors(params.region, params.shard, params.account, params.project, params.task, 100).then((res) => {
          this.ancestors = res.ancestors
          this.moreAncestors = res.more
        })
        mapi.v1.task.getChildren(params.region, params.shard, params.account, params.project, params.task, null, 100).then((res) => {
          this.children = res.children
          this.moreChildren = res.more
        })
        mapi.v1.timeLog.get(params.region, params.shard, params.account, params.project, params.task, null, null, true, null, 100).then((res) => {
          this.timeLogs = res.timeLogs
          this.moreTimeLogs = res.more
          this.loading = false
        })
        mapi.sendMGet()
      },
      loadMore () {
        if (!this.loading) {
          let params = router.currentRoute.params
          this.loading = true
          if (this.task.isAbstract) {
            let after = null
            if (this.children.length > 0) {
              after = this.children[this.children.length - 1].id
            }
            api.v1.task.getChildren(params.region, params.shard, params.account, params.project, params.task, after, 100).then((res) => {
              this.loading = false
              res.children.forEach((child) => {
                this.children.push(child)
              })
              this.moreChildren = res.more
            })
          } else {
            let after = null
            if (this.timeLogs.length > 0) {
              after = this.timeLogs[this.timeLogs.length - 1].id
            }
            api.v1.timeLog.get(params.region, params.shard, params.account, params.project, params.task, null, null, true, after, 100).then((res) => {
              this.loading = false
              res.timeLogs.forEach((tl) => {
                this.timeLogs.push(tl)
              })
              this.moreTimeLogs = res.more
            })
          }
        }
      },
      goToTask (task) {
        let params = router.currentRoute.params
        router.push('/app/region/' + params.region + '/shard/' + params.shard + /account/ + params.account + /project/ + params.project + /task/ + task.id)
        this.init()
      },
      toggleCreateTaskForm () {
        if (!this.creatingTask) {
          this.createTaskDialog = !this.createTaskDialog
          this.createTaskName = ''
          this.createTaskDescription = null
          this.createTaskIsParallel = true
          this.createTaskRemainingTime = null
        }
      },
      toggleCreateTimeLogForm () {
        if (!this.creatingTimeLog) {
          this.createTimeLogDialog = !this.createTimeLogDialog
          this.createTimeLogDuration = null
          this.createTimeLogRemainingTime = null
          this.createTimeLogNote = null
        }
      },
      createTask () {
        this.creatingTask = true
        let params = router.currentRoute.params
        let description = null
        if (this.createTaskDescription && this.createTaskDescription.length) {
          description = this.createTaskDescription
        }
        let previousSibling = null
        if (this.children.length > 0) {
          previousSibling = this.children[this.children.length - 1].id
        }
        let isParallel = this.createTaskIsParallel
        if (!this.createTaskIsAbstract) {
          isParallel = null
        }
        let remainingTime = parseInt(this.createTaskRemainingTime)
        if (this.createTaskIsAbstract) {
          remainingTime = null
        }
        let myId = null
        api.v1.centralAccount.getMe().then((me) => {
          if (!this.createTaskIsAbstract) {
            myId = me.id
          }
          api.v1.task.create(params.region, params.shard, params.account, params.project, params.task, previousSibling, this.createTaskName, description, this.createTaskIsAbstract, isParallel, myId, remainingTime).then((newTask) => {
            this.creatingTask = false
            this.toggleCreateTaskForm()
            this.children.push(newTask)
            this.task.childCount++
            this.task.descendantCount++
            if (remainingTime) {
              this.task.totalRemainingTime += remainingTime
              if (this.task.isParallel) {
                if (this.task.minimumRemainingTime < remainingTime) {
                  this.task.minimumRemainingTime = remainingTime
                }
              } else {
                this.task.minimumRemainingTime += remainingTime
              }
            }
          })
        })
      },
      createTimeLog () {
        if (this.createTimeLogDuration !== null && parseInt(this.createTimeLogDuration) > 0) {
          this.creatingTimeLog = true
          let params = router.currentRoute.params
          let note = null
          if (this.createTimeLogNote !== null && this.createTimeLogNote.length > 0) {
            note = this.createTimeLogNote
          }
          let resHandler = (newTimeLog) => {
            this.creatingTimeLog = false
            this.timeLogs.push(newTimeLog)
            this.task.totalLoggedTime += parseInt(this.createTimeLogDuration)
            if (this.createTimeLogRemainingTime !== null && parseInt(this.createTimeLogRemainingTime) >= 0) {
              this.task.totalRemainingTime = parseInt(this.createTimeLogRemainingTime)
            }
            this.toggleCreateTimeLogForm()
          }
          if (this.createTimeLogRemainingTime !== null && this.createTimeLogRemainingTime >= 0) {
            api.v1.timeLog.createAndSetRemainingTime(params.region, params.shard, params.account, params.project, params.task, parseInt(this.createTimeLogRemainingTime), parseInt(this.createTimeLogDuration), note).then(resHandler)
          } else {
            api.v1.timeLog.create(params.region, params.shard, params.account, params.project, params.task, this.createTimeLogDuration, note).then(resHandler)
          }
        }
      }
    }
  }
</script>

<style scoped lang="scss">
  h3, span {
    display: inline-flex;
  }
</style>

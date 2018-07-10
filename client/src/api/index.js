import axios from 'axios'

export const cnst = {
  regions: {
    central: 'central', // Central directory
    use: 'use', // US East
    usw: 'usw', // US West
    euw: 'euw', // EU West
    asp: 'asp', // Asia Pacific
    aus: 'aus' // Australia
  },
  env: {
    lcl: 'lcl', // local
    dev: 'dev', // develop
    stg: 'stg', // staging
    pro: 'pro' // production
  },
  theme: {
    light: 0,
    dark: 1,
    colorBlind: 2
  },
  accountRole: {
    owner: 0,
    admin: 1,
    memberOfAllProjects: 2,
    memberOfOnlySpecificProjects: 3
  },
  projectRole: {
    admin: 0,
    writer: 1,
    reader: 2
  },
  sortBy: {
    name: 'name',
    displayName: 'displayname', // only used for users
    createdOn: 'createdon',
    // only used for projects
    startOn: 'starton',
    dueOn: 'dueon'
  }
}

let memCache = {}

let newApi
newApi = (opts) => {
  let isMDoApi = opts.isMDoApi
  let mDoApiRegion = opts.mDoApiRegion
  let mDoSending = false
  let mDoSent = false
  let awaitingMDoList = []
  let doReq = (region, path, args) => {
    if (args && typeof args.shard === 'string') {
      args.shard = parseInt(args.shard, 10)
    }
    if (!isMDoApi || (isMDoApi && mDoSending && !mDoSent)) {
      return axios({
        method: 'post',
        url: path + '?region=' + region,
        data: args,
        headers: {
          'X-Client': 'web'
        }
      }).then((res) => {
        return res.data
      })
    } else if (isMDoApi && !mDoSending && !mDoSent) {
      if (region !== mDoApiRegion) {
        throw new Error('invalid mdo call, all calls must be to the same region')
      }
      let awaitingMDoObj = {
        region: region,
        path: path,
        args: args,
        resolve: null,
        reject: null
      }
      awaitingMDoList.push(awaitingMDoObj)
      return new Promise((resolve, reject) => {
        awaitingMDoObj.resolve = resolve
        awaitingMDoObj.reject = reject
      })
    } else {
      throw new Error('invalid get call, use the default api object or a new mdo instance from api.newMDoApi()')
    }
  }

  return {
    newMDoApi: (region) => {
      return newApi({isMDoApi: true, mDoApiRegion: region})
    },
    sendMDo: () => {
      if (!isMDoApi) {
        throw new Error('MDoes must be made from the api instance returned from api.newMDoApi()')
      } else if (mDoSending || mDoSent) {
        throw new Error('each MDo must be started with a fresh api.newMDoApi(), once used that same instance cannot be reused')
      }
      mDoSending = true
      let asyncIndividualPromisesReady
      asyncIndividualPromisesReady = (resolve) => {
        let ready = true
        for (let i = 0, l = awaitingMDoList.length; i < l; i++) {
          if (awaitingMDoList[i].resolve === null) {
            ready = false
            setTimeout(asyncIndividualPromisesReady, 0, resolve)
          }
        }
        if (ready) {
          resolve()
        }
      }
      let mDoComplete = false
      let mDoCompleterFunc
      mDoCompleterFunc = (resolve) => {
        if (mDoComplete) {
          resolve()
        } else {
          setTimeout(mDoCompleterFunc, 0, resolve)
        }
      }
      new Promise(asyncIndividualPromisesReady).then(() => {
        let mDoObj = {}
        for (let i = 0, l = awaitingMDoList.length; i < l; i++) {
          let key = '' + i
          mDoObj[key] = {
            region: awaitingMDoList[i].region,
            path: awaitingMDoList[i].path,
            args: awaitingMDoList[i].args
          }
        }
        doReq(mDoApiRegion, '/api/mdo', mDoObj).then((res) => {
          mDoSending = false
          mDoSent = true
          for (let i = 0, l = awaitingMDoList.length; i < l; i++) {
            let key = '' + i
            if (res[key].code === 200) {
              awaitingMDoList[i].resolve(res[key].body)
            } else {
              awaitingMDoList[i].reject(res[key])
            }
          }
          mDoComplete = true
        }).catch((error) => {
          mDoComplete = true
          mDoSending = false
          mDoSent = true
          for (let i = 0, l = awaitingMDoList.length; i < l; i++) {
            awaitingMDoList[i].reject(error)
          }
        })
      })
      return new Promise(mDoCompleterFunc)
    },
    logout: () => {
      memCache = {}
      return doReq(cnst.regions.central, '/api/logout')
    },
    v1: {
      centralAccount: {
        register: (name, email, pwd, region, language, displayName, theme) => {
          return doReq(cnst.regions.central, '/api/v1/centralAccount/register', {name, email, pwd, region, language, displayName, theme})
        },
        resendActivationEmail: (email) => {
          return doReq(cnst.regions.central, '/api/v1/centralAccount/resendActivationEmail', {email})
        },
        activate: (email, activationCode) => {
          return doReq(cnst.regions.central, '/api/v1/centralAccount/activate', {email, activationCode})
        },
        authenticate: (email, pwdTry) => {
          return doReq(cnst.regions.central, '/api/v1/centralAccount/authenticate', {email, pwdTry}).then((res) => {
            memCache.me = res.me
            memCache[memCache.me.id] = memCache.me
            return res
          })
        },
        confirmNewEmail: (currentEmail, newEmail, confirmationCode) => {
          return doReq(cnst.regions.central, '/api/v1/centralAccount/confirmNewEmail', {currentEmail, newEmail, confirmationCode})
        },
        resetPwd: (email) => {
          return doReq(cnst.regions.central, '/api/v1/centralAccount/resetPwd', {email})
        },
        setNewPwdFromPwdReset: (newPwd, email, resetPwdCode) => {
          return doReq(cnst.regions.central, '/api/v1/centralAccount/setNewPwdFromPwdReset', {newPwd, email, resetPwdCode})
        },
        getAccount: (name) => {
          return doReq(cnst.regions.central, '/api/v1/centralAccount/getAccount', {name})
        },
        getAccounts: (accounts) => {
          return doReq(cnst.regions.central, '/api/v1/centralAccount/getAccounts', {accounts})
        },
        searchAccounts: (nameOrDisplayNameStartsWith) => {
          return doReq(cnst.regions.central, '/api/v1/centralAccount/searchAccounts', {nameOrDisplayNameStartsWith})
        },
        searchPersonalAccounts: (nameOrDisplayNameStartsWith) => {
          return doReq(cnst.regions.central, '/api/v1/centralAccount/namesearchPersonalAccounts', {nameOrDisplayNameStartsWith})
        },
        getMe: () => {
          if (memCache.me) {
            return new Promise((resolve) => {
              resolve(memCache.me)
            })
          }
          return doReq(cnst.regions.central, '/api/v1/centralAccount/getMe').then((res) => {
            memCache.me = res
            return res
          })
        },
        setMyPwd: (oldPwd, newPwd) => {
          return doReq(cnst.regions.central, '/api/v1/centralAccount/setMyPwd', {oldPwd, newPwd})
        },
        setMyEmail: (newEmail) => {
          return doReq(cnst.regions.central, '/api/v1/centralAccount/setMyEmail', {newEmail})
        },
        resendMyNewEmailConfirmationEmail: () => {
          return doReq(cnst.regions.central, '/api/v1/centralAccount/resendMyNewEmailConfirmationEmail')
        },
        setAccountName: (account, newName) => {
          return doReq(cnst.regions.central, '/api/v1/centralAccount/setAccountName', {account, newName})
        },
        setAccountDisplayName: (account, newDisplayName) => {
          return doReq(cnst.regions.central, '/api/v1/centralAccount/setAccountDisplayName', {account, newDisplayName})
        },
        setAccountAvatar: (account, avatar) => {
          let data = new FormData()
          data.append('account', account)
          if (avatar) {
            data.append('avatar', avatar, '')
          }
          return doReq(cnst.regions.central, '/api/v1/centralAccount/setAccountAvatar', data)
        },
        migrateAccount: (account, newRegion) => {
          return doReq(cnst.regions.central, '/api/v1/centralAccount/migrateAccount', {account, newRegion})
        },
        createAccount: (name, region, displayName) => {
          return doReq(cnst.regions.central, '/api/v1/centralAccount/createAccount', {name, region, displayName})
        },
        getMyAccounts: (after, limit) => {
          return doReq(cnst.regions.central, '/api/v1/centralAccount/getMyAccounts', {after, limit})
        },
        deleteAccount: (account) => {
          return doReq(cnst.regions.central, '/api/v1/centralAccount/deleteAccount', {account})
        },
        addMembers: (account, newMembers) => {
          return doReq(cnst.regions.central, '/api/v1/centralAccount/addMembers', {account, newMembers})
        },
        removeMembers: (account, existingMembers) => {
          return doReq(cnst.regions.central, '/api/v1/centralAccount/removeMembers', {account, existingMembers})
        }
      },
      account: {
        edit: (region, shard, account, fields) => {
          return doReq(region, '/api/v1/account/edit', {shard, account, fields})
        },
        get: (region, shard, account) => {
          return doReq(region, '/api/v1/account/get', {shard, account})
        },
        setMemberRole: (region, shard, account, member, role) => {
          return doReq(region, '/api/v1/account/setMemberRole', {shard, account, member, role})
        },
        getMembers: (region, shard, account, role, nameContains, after, limit) => {
          return doReq(region, '/api/v1/account/getMembers', {shard, account, role, nameContains, after, limit})
        },
        getActivities: (region, shard, account, item, member, occurredAfter, occurredBefore, limit) => {
          return doReq(region, '/api/v1/account/getActivities', {shard, account, item, member, occurredAfter, occurredBefore, limit})
        },
        getMe: (region, shard, account) => {
          return doReq(region, '/api/v1/account/getMe', {shard, account})
        }
      },
      project: {
        create: (region, shard, account, name, description, hoursPerDay, daysPerWeek, startOn, dueOn, isParallel, isPublic, members) => {
          return doReq(region, '/api/v1/project/create', {shard, account, name, description, hoursPerDay, daysPerWeek, startOn, dueOn, isParallel, isPublic, members})
        },
        edit: (region, shard, account, project, fields) => {
          return doReq(region, '/api/v1/project/edit', {shard, account, project, fields})
        },
        get: (region, shard, account, project) => {
          return doReq(region, '/api/v1/project/get', {shard, account, project})
        },
        getSet: (region, shard, account, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortAsc, after, limit) => {
          return doReq(region, '/api/v1/project/getSet', {shard, account, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortAsc, after, limit})
        },
        delete: (region, shard, account, project) => {
          return doReq(region, '/api/v1/project/delete', {shard, account, project})
        },
        addMembers: (region, shard, account, project, members) => {
          return doReq(region, '/api/v1/project/addMembers', {shard, account, project, members})
        },
        setMemberRole: (region, shard, account, project, member, role) => {
          return doReq(region, '/api/v1/project/setMemberRole', {shard, account, project, member, role})
        },
        removeMembers: (region, shard, account, project, members) => {
          return doReq(region, '/api/v1/project/removeMembers', {shard, account, project, members})
        },
        getMembers: (region, shard, account, project, role, nameOrDisplayNameContains, after, limit) => {
          return doReq(region, '/api/v1/project/getMembers', {shard, account, project, role, nameOrDisplayNameContains, after, limit})
        },
        getMe: (region, shard, account, project) => {
          return doReq(region, '/api/v1/project/getMembers', {shard, account, project})
        },
        getActivities: (region, shard, account, project, item, member, occurredAfter, occurredBefore, limit) => {
          return doReq(region, '/api/v1/project/getActivities', {shard, account, project, item, member, occurredAfter, occurredBefore, limit})
        }
      },
      task: {
        create: (region, shard, account, project, parent, previousSibling, name, description, isAbstract, isParallel, member, totalRemainingTime) => {
          return doReq(region, '/api/v1/task/create', {shard, account, project, parent, previousSibling, name, description, isAbstract, isParallel, member, totalRemainingTime})
        },
        setName: (region, shard, account, project, task, name) => {
          return doReq(region, '/api/v1/task/setName', {shard, account, project, task, name})
        },
        setDescription: (region, shard, account, project, task, description) => {
          return doReq(region, '/api/v1/task/setDescription', {shard, account, project, task, description})
        },
        setIsParallel: (region, shard, account, project, task, isParallel) => {
          return doReq(region, '/api/v1/task/setDescription', {shard, account, project, task, isParallel})
        },
        setMember: (region, shard, account, project, task, member) => {
          return doReq(region, '/api/v1/task/setMember', {shard, account, project, task, member})
        },
        setRemainingTime: (region, shard, account, project, task, remainingTime) => {
          return doReq(region, '/api/v1/task/setremainingTime', {shard, account, project, task, remainingTime})
        },
        move: (region, shard, account, project, task, parent, nextSibling) => {
          return doReq(region, '/api/v1/task/move', {shard, account, project, task, parent, nextSibling})
        },
        delete: (region, shard, account, project, task) => {
          return doReq(region, '/api/v1/task/delete', {shard, account, project, task})
        },
        get: (region, shard, account, project, task) => {
          return doReq(region, '/api/v1/task/get', {shard, account, project, task})
        },
        getChildren: (region, shard, account, project, parent, fromSibling, limit) => {
          return doReq(region, '/api/v1/task/getChildren', {shard, account, project, parent, fromSibling, limit})
        },
        getAncestors: (region, shard, account, project, child, limit) => {
          return doReq(region, '/api/v1/task/getAncestors', {shard, account, project, child, limit})
        }
      },
      timeLog: {
        create: (region, shard, account, project, task, duration, note) => {
          return doReq(region, '/api/v1/timeLog/create', {shard, account, project, task, duration, note})
        },
        createAndSetRemainingTime: (region, shard, account, project, task, remainingTime, duration, note) => {
          return doReq(region, '/api/v1/timeLog/createAndSetRemainingTime', {shard, account, project, task, remainingTime, duration, note})
        },
        setDuration: (region, shard, account, project, timeLog, duration) => {
          return doReq(region, '/api/v1/timeLog/setDuration', {shard, account, project, timeLog, duration})
        },
        setNote: (region, shard, account, project, timeLog, note) => {
          return doReq(region, '/api/v1/timeLog/setNote', {shard, account, project, timeLog, note})
        },
        delete: (region, shard, account, project, timeLog) => {
          return doReq(region, '/api/v1/timeLog/delete', {shard, account, project, timeLog})
        },
        get: (region, shard, account, project, task, member, timeLog, sortAsc, after, limit) => {
          return doReq(region, '/api/v1/timeLog/get', {shard, account, project, task, member, timeLog, sortAsc, after, limit})
        }
      }
    }
  }
}

export default newApi({isMDoApi: false})

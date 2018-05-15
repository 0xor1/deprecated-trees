/**
 * IMPORTANT: this file should only be altered by backend api developers
 * **/

import axios from 'axios'

let newApi
newApi = (opts) => {
  let isMGetApi = opts.isMGetApi
  let mGetSending = false
  let mGetSent = false
  let awaitingMGetList = []

  let doReq = (axiosConfig) => {
    axiosConfig['X-Client'] = 'web'
    return axios(axiosConfig)
  }

  let get = (path, data) => {
    let url = path
    if (typeof data === 'object') {
      url = url + '?args=' + encodeURIComponent(JSON.stringify(data))
    }
    if (!isMGetApi || (mGetSending && !mGetSent)) {
      return doReq({
        method: 'get',
        url: url
      })
    } else if (isMGetApi && !mGetSending && !mGetSent) {
      let awaitingMGetObj = {
        url: url,
        resolve: null,
        reject: null
      }
      awaitingMGetList.push(awaitingMGetObj)
      return new Promise((resolve, reject) => {
        awaitingMGetObj.resolve = resolve
        awaitingMGetObj.reject = reject
      })
    } else {
      throw new Error('invalid get call, use the default api object or a new mget instance from api.newMGetApi()')
    }
  }

  let post = (path, data) => {
    doReq({
      method: 'post',
      url: path,
      data: data
    })
  }

  return {
    newMGetApi: () => {
      return newApi({isMGetApi: true})
    },
    sendMGet: () => {
      if (!isMGetApi) {
        throw new Error('MGets must be made from the api instance returned from api.newMGetApi()')
      } else if (mGetSending || mGetSent) {
        throw new Error('each MGet must be started with a fresh api.newMGetApi(), once used that same instance cannot be reused')
      } else if (awaitingMGetList <= 1) {
        throw new Error('sending MGet requests should only be done with more than 1 request, otherwise just use a regular get')
      }
      mGetSending = true
      let asyncIndividualPromisesReady
      asyncIndividualPromisesReady = (resolve) => {
        for (let i = 0, l = awaitingMGetList.length; i < l; i++) {
          if (awaitingMGetList[i].resolve === null) {
            setTimeout(asyncIndividualPromisesReady, 0, resolve)
          }
        }
        resolve()
      }
      new Promise(asyncIndividualPromisesReady).then(() => {
        let mgetObj = {}
        for (let i = 0, l = awaitingMGetList.length; i < l; i++) {
          let key = '' + i
          mgetObj[key] = awaitingMGetList[i].url
        }
        get('/api/mget', mgetObj).then((res) => {
          mGetSending = false
          mGetSent = true
          for (let i = 0, l = awaitingMGetList.length; i < l; i++) {
            let key = '' + i
            if (res.data[key].code === 200) {
              awaitingMGetList[i].resolve(res.data[key].body)
            } else {
              awaitingMGetList[i].reject(res.data[key].body)
            }
          }
        }).catch((error) => {
          mGetSending = false
          mGetSent = true
          for (let i = 0, l = awaitingMGetList.length; i < l; i++) {
            awaitingMGetList[i].reject(error)
          }
        })
      })
    },
    v1: {
      centralAccount: {
        getRegions: () => {
          return get('/api/v1/centralAccount/getRegions')
        },
        register: (name, email, pwd, region, language, displayName, theme) => {
          return post('/api/v1/centralAccount/register', {name: name, email: email, pwd: pwd, region: region, language: language, displayName: displayName, theme: theme})
        },
        resendActivationEmail: (email) => {
          return post('/api/v1/centralAccount/resendActivationEmail', {email: email})
        },
        activate: (email, activationCode) => {
          return post('/api/v1/centralAccount/activate', {email: email, activationCode: activationCode})
        },
        authenticate: (email, pwd) => {
          return post('/api/v1/centralAccount/authenticate', {email: email, pwd: pwd})
        },
        confirmNewEmail: (currentEmail, newEmail, confirmationCode) => {
          return post('/api/v1/centralAccount/confirmNewEmail', {currentEmail: currentEmail, newEmail: newEmail, confirmationCode: confirmationCode})
        },
        resetPwd: (email) => {
          return post('/api/v1/centralAccount/resetPwd', {email: email})
        },
        setNewPwdFromPwdReset: (newPwd, email, resetPwdCode) => {
          return post('/api/v1/centralAccount/setNewPwdFromPwdReset', {newPwd: newPwd, email: email, resetPwdCode: resetPwdCode})
        },
        getAccount: (name) => {
          return get('/api/v1/centralAccount/getAccount', {name: name})
        },
        getAccounts: (accounts) => {
          return get('/api/v1/centralAccount/getAccounts', {accounts: accounts})
        },
        searchAccounts: (nameOrDisplayNameStartsWith) => {
          return get('/api/v1/centralAccount/searchAccounts', {nameOrDisplayNameStartsWith: nameOrDisplayNameStartsWith})
        },
        searchPersonalAccounts: (nameOrDisplayNameStartsWith) => {
          return get('/api/v1/centralAccount/namesearchPersonalAccounts', {nameOrDisplayNameStartsWith: nameOrDisplayNameStartsWith})
        }
      },
      account: {},
      project: {},
      task: {},
      timeLog: {}
    }
  }
}

export default newApi({isMGetApi: false})

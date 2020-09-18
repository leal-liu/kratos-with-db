#!/usr/bin/env python3


import os
import shutil
import subprocess
import sys

chain_id = "kratos"
moniker = "db_backend"
master_ip = "121.89.211.107"
seed_id = "c3a78fb5ed3afe777d4debc66a1ef2fddd340476"
genesis_json = """
{
  "genesis_time": "2020-09-15T08:49:52.832713046Z",
  "chain_id": "kratos",
  "consensus_params": {
    "block": {
      "max_bytes": "22020096",
      "max_gas": "-1",
      "time_iota_ms": "1000"
    },
    "evidence": {
      "max_age_num_blocks": "100000",
      "max_age_duration": "172800000000000"
    },
    "validator": {
      "pub_key_types": [
        "ed25519"
      ]
    }
  },
  "app_hash": "",
  "app_state": {
    "mint": {
      "minter": {
        "inflation": "0.140000000000000000",
        "annual_provisions": "0.000000000000000000"
      },
      "params": {
        "mint_denom": "kratos/kts",
        "inflation_rate_change": "0.140000000000000000",
        "inflation_max": "0.210000000000000000",
        "inflation_min": "0.080000000000000000",
        "goal_bonded": "0.670000000000000000",
        "blocks_per_year": "10519200"
      }
    },
    "kudistribution": {
      "params": {
        "community_tax": "0.020000000000000000",
        "base_proposer_reward": "0.010000000000000000",
        "bonus_proposer_reward": "0.040000000000000000",
        "withdraw_addr_enabled": true
      },
      "fee_pool": {
        "community_pool": []
      },
      "delegator_withdraw_infos": [],
      "previous_proposer": "",
      "outstanding_rewards": [],
      "validator_accumulated_commissions": [],
      "validator_historical_rewards": [],
      "validator_current_rewards": [],
      "delegator_starting_infos": [],
      "validator_slash_events": []
    },
    "genutil": {
      "gentxs": [
        {
          "type": "kuchain/Tx",
          "value": {
            "msg": [
              {
                "type": "kuchain/KuMsgCreateValidator",
                "value": {
                  "KuMsg": {
                    "auth": [
                      "kratos1s0g2w3mz9se507pknahltl53rvq024pxvlj4wq"
                    ],
                    "from": "",
                    "to": "",
                    "amount": [],
                    "router": "kustaking",
                    "action": "create@staking",
                    "data": "pQH3Pbq/CggKBmtyYXRvcxISMTAwMDAwMDAwMDAwMDAwMDAwGhMKEQEBBi0gVD0wAAAAAAAAAAAAIhcKFQKD0KdHYiwzR/g2n2/1/pEbAPVUJipTa3JhdG9zdmFsY29uc3B1YjF6Y2pkdWVwcXF3a3R1cHg5MjYydDBjazVndWdwZ2d0OG1rc2hkcXE3Mzk5M210aGVzNWUybmw4ZjcycXNzZHJrcjA="
                  }
                }
              },
              {
                "type": "kuchain/KuMsgDelegate",
                "value": {
                  "KuMsg": {
                    "auth": [
                      "kratos1s0g2w3mz9se507pknahltl53rvq024pxvlj4wq"
                    ],
                    "from": "kratos",
                    "to": "kustaking",
                    "amount": [
                      {
                        "denom": "kratos/kts",
                        "amount": "1000000000000000000"
                      }
                    ],
                    "router": "kustaking",
                    "action": "delegate",
                    "data": "Ue+AYjUKEwoRAQEGLSBUPTAAAAAAAAAAAAASEwoRAQEGLSBUPTAAAAAAAAAAAAAaIQoKa3JhdG9zL2t0cxITMTAwMDAwMDAwMDAwMDAwMDAwMA=="
                  }
                }
              }
            ],
            "fee": {
              "amount": [
                {
                  "denom": "kratos/kts",
                  "amount": "2000"
                }
              ],
              "gas": "200000",
              "payer": "kratos"
            },
            "signatures": [
              {
                "pub_key": {
                  "type": "tendermint/PubKeySecp256k1",
                  "value": "A+dVi+zdKndpedI6AIEQhOg6eyxHHKU4krg5xu569W/3"
                },
                "signature": "ixQtOd0pqu7XgZAjTrhkfnhYBNTluY2hOHY/p+iCGJx312O/xPwcjP08onGbM/O5BIYVDaagthhubfF8DjPQcQ=="
              }
            ],
            "memo": "ba4684e9bdba4cb1be1cecb238da4fbbfd491626@172.26.37.12:26656"
          }
        }
      ]
    },
    "supply": {
      "supply": []
    },
    "kustaking": {
      "params": {
        "unbonding_time": "1209600000000000",
        "max_validators": 33,
        "max_entries": 7,
        "bond_denom": "kratos/kts"
      },
      "last_total_power": "0",
      "last_validator_powers": null,
      "validators": null,
      "delegations": null,
      "unbonding_delegations": null,
      "redelegations": null,
      "exported": false
    },
    "plugin": {
      "type": "plugin/genesisState",
      "value": {}
    },
    "kuslashing": {
      "params": {
        "signed_blocks_window": "100",
        "min_signed_per_window": "0.500000000000000000",
        "downtime_jail_duration": "600000000000",
        "slash_fraction_double_sign": "0.050000000000000000",
        "slash_fraction_downtime": "0.000100000000000000"
      },
      "signing_infos": {},
      "missed_blocks": {}
    },
    "account": {
      "accounts": [
        {
          "type": "kuchain/Account",
          "value": {
            "id": "kratos",
            "account_number": "1",
            "auths": [
              {
                "name": "root",
                "address": "kratos1s0g2w3mz9se507pknahltl53rvq024pxvlj4wq"
              }
            ]
          }
        },
        {
          "type": "kuchain/Account",
          "value": {
            "id": "kratos1s0g2w3mz9se507pknahltl53rvq024pxvlj4wq",
            "account_number": "2",
            "auths": null
          }
        },
        {
          "type": "kuchain/Account",
          "value": {
            "id": "initial@kratos",
            "account_number": "3",
            "auths": [
              {
                "name": "root",
                "address": "kratos1n0r4jp2m7ea7qm8lfe6eh8n7veflvrla537p59"
              }
            ]
          }
        },
        {
          "type": "kuchain/Account",
          "value": {
            "id": "foundation@kratos",
            "account_number": "4",
            "auths": [
              {
                "name": "root",
                "address": "kratos1pd82nlsvpty9k4x6lks62a3ch0z4dftreyhkz5"
              }
            ]
          }
        }
      ]
    },
    "kuevidence": {
      "params": {
        "max_evidence_age": "120000000000",
        "double_sign_jail_duration": "1209600000000000"
      },
      "evidence": []
    },
    "kugov": {
      "starting_proposal_id": "1",
      "deposits": null,
      "votes": null,
      "proposals": null,
      "deposit_params": {
        "min_deposit": [
          {
            "denom": "kratos/kts",
            "amount": "500000000000000000000"
          }
        ],
        "max_deposit_period": "1209600000000000"
      },
      "voting_params": {
        "voting_period": "1209600000000000"
      },
      "tally_params": {
        "quorum": "0.334000000000000000",
        "threshold": "0.500000000000000000",
        "veto": "0.334000000000000000",
        "emergency": "0.667000000000000000",
        "max_punish_period": "604800000000000",
        "slash_fraction": "0.000100000000000000"
      }
    },
    "asset": {
      "type": "asset/genesisState",
      "value": {
        "genesisAssets": [
          {
            "type": "asset/genesisAsset",
            "value": {
              "id": "kratos",
              "coins": [
                {
                  "denom": "kratos/kts",
                  "amount": "25000000000000000000"
                }
              ]
            }
          },
          {
            "type": "asset/genesisAsset",
            "value": {
              "id": "kratos1s0g2w3mz9se507pknahltl53rvq024pxvlj4wq",
              "coins": [
                {
                  "denom": "kratos/kts",
                  "amount": "25000000000000000000"
                }
              ]
            }
          },
          {
            "type": "asset/genesisAsset",
            "value": {
              "id": "initial@kratos",
              "coins": [
                {
                  "denom": "kratos/kts",
                  "amount": "100000000000000000000000000"
                }
              ]
            }
          },
          {
            "type": "asset/genesisAsset",
            "value": {
              "id": "foundation@kratos",
              "coins": [
                {
                  "denom": "kratos/kts",
                  "amount": "40000000000000000000000000"
                }
              ]
            }
          }
        ],
        "genesisCoins": [
          {
            "type": "asset/genesisCoin",
            "value": {
              "creator": "kratos",
              "symbol": "kts",
              "maxSupply": {
                "denom": "kratos/kts",
                "amount": "0"
              },
              "description": "core token for kratos chain"
            }
          }
        ]
      }
    },
    "kuparams": null
  }
}
"""


def replace_all(text, *argv):
    """
    替换字符串
    :param text:
    :param argv:
    :return:
    """
    import re
    for i in range(0, len(argv), 2):
        text = re.sub(argv[i], argv[i + 1], text)
    return text


def init():
    """
    初始化
    :return:
    """
    # 设置环境变量
    bin_dir = 'build'
    if os.path.exists(bin_dir):
        path = os.getenv('PATH')
        os.putenv('PATH', '{}:{}'.format(path, os.path.join(os.getcwd(), bin_dir)))

    # 检查Home是否存在
    chain_home = os.path.expanduser('~/.kucd')
    if os.path.exists(chain_home):
        return

    # 初始化链
    cmd = "kucd init --chain-id={} {}".format(chain_id, moniker)
    subprocess.check_call(cmd, shell=True)

    # 替换genesis.json文件
    genesis_json_path = os.path.expanduser('~/.kucd/config/genesis.json')
    with open(genesis_json_path, 'w') as f:
        f.write(genesis_json)

    # 替换配置
    config_toml_path = os.path.expanduser('~/.kucd/config/config.toml')
    with open(config_toml_path) as f:
        text = f.read()
    text = replace_all(text,
                       '0.0.0.0:26656', '0.0.0.0:36656',
                       '127.0.0.1:26657', '0.0.0.0:36657',
                       'persistent_peers = ""', 'persistent_peers = "{}@{}:36656"'.format(seed_id, master_ip),
                       'private_peer_ids = ""', 'private_peer_ids = "{}"'.format(seed_id))
    with open(config_toml_path, 'w') as f:
        f.write(text)


def main(argv):
    # 判断是否只初始化
    just_init = False
    if 1 < len(argv) and 'init' == argv[1]:
        just_init = True
    if just_init:
        chain_home = os.path.expanduser('~/.kucd')
        shutil.rmtree(chain_home, ignore_errors=True)
    # 初始化
    init()
    if just_init:
        return
    # 启动进程
    cmd = 'kucd start --plugin-cfg plugins.json'
    subprocess.check_call(cmd, shell=True)


if __name__ == '__main__':
    main(sys.argv)

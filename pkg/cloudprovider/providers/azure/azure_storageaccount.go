/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package azure

import (
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/arm/storage"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/golang/glog"
)

type accountWithLocation struct {
	Name, StorageType, Location string
}

// getStorageAccounts gets name, type, location of all storage accounts in a resource group which matches matchingAccountType, matchingLocation
func (az *Cloud) getStorageAccounts(matchingAccountType, resourceGroup, matchingLocation string) ([]accountWithLocation, error) {
	az.operationPollRateLimiter.Accept()
	glog.V(10).Infof("StorageAccountClient.ListByResourceGroup(%v): start", resourceGroup)
	result, err := az.StorageAccountClient.ListByResourceGroup(resourceGroup)
	glog.V(10).Infof("StorageAccountClient.ListByResourceGroup(%v): end", resourceGroup)
	if err != nil {
		return nil, err
	}
	if result.Value == nil {
		return nil, fmt.Errorf("unexpected error when listing storage accounts from resource group %s", resourceGroup)
	}

	accounts := []accountWithLocation{}
	for _, acct := range *result.Value {
		if acct.Name != nil && acct.Location != nil && acct.Sku != nil {
			storageType := string((*acct.Sku).Name)
			if matchingAccountType != "" && !strings.EqualFold(matchingAccountType, storageType) {
				continue
			}

			location := *acct.Location
			if matchingLocation != "" && !strings.EqualFold(matchingLocation, location) {
				continue
			}
			accounts = append(accounts, accountWithLocation{Name: *acct.Name, StorageType: storageType, Location: location})
		}
	}

	return accounts, nil
}

// getStorageAccesskey gets the storage account access key
func (az *Cloud) getStorageAccesskey(account, resourceGroup string) (string, error) {
	az.operationPollRateLimiter.Accept()
	glog.V(10).Infof("StorageAccountClient.ListKeys(%q): start", account)
	result, err := az.StorageAccountClient.ListKeys(resourceGroup, account)
	glog.V(10).Infof("StorageAccountClient.ListKeys(%q): end", account)
	if err != nil {
		return "", err
	}
	if result.Keys == nil {
		return "", fmt.Errorf("empty keys")
	}

	for _, k := range *result.Keys {
		if k.Value != nil && *k.Value != "" {
			v := *k.Value
			if ind := strings.LastIndex(v, " "); ind >= 0 {
				v = v[(ind + 1):]
			}
			return v, nil
		}
	}
	return "", fmt.Errorf("no valid keys")
}

// ensureStorageAccount search storage account, create one storage account(with genAccountNamePrefix) if not found, return accountName, accountKey
func (az *Cloud) ensureStorageAccount(accountName, accountType, resourceGroup, location, genAccountNamePrefix string) (string, string, error) {
	if len(accountName) == 0 {
		// find a storage account that matches accountType
		accounts, err := az.getStorageAccounts(accountType, resourceGroup, location)
		if err != nil {
			return "", "", fmt.Errorf("could not list storage accounts for account type %s: %v", accountType, err)
		}

		if len(accounts) > 0 {
			accountName = accounts[0].Name
			glog.V(4).Infof("found a matching account %s type %s location %s", accounts[0].Name, accounts[0].StorageType, accounts[0].Location)
		}

		if len(accountName) == 0 {
			// not found a matching account, now create a new account in current resource group
			accountName = generateStorageAccountName(genAccountNamePrefix)
			if location == "" {
				location = az.Location
			}
			if accountType == "" {
				accountType = defaultStorageAccountType
			}

			glog.V(2).Infof("azure - no matching account found, begin to create a new account %s in resource group %s, location: %s, accountType: %s",
				accountName, resourceGroup, location, accountType)
			cp := storage.AccountCreateParameters{
				Sku:  &storage.Sku{Name: storage.SkuName(accountType)},
				Tags: &map[string]*string{"created-by": to.StringPtr("azure")},
				AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{EnableHTTPSTrafficOnly: to.BoolPtr(true)},
				Location: &location}
			cancel := make(chan struct{})

			_, errchan := az.StorageAccountClient.Create(resourceGroup, accountName, cp, cancel)
			err := <-errchan
			if err != nil {
				return "", "", fmt.Errorf(fmt.Sprintf("Failed to create storage account %s, error: %s", accountName, err))
			}
		}
	}

	// find the access key with this account
	accountKey, err := az.getStorageAccesskey(accountName, resourceGroup)
	if err != nil {
		return "", "", fmt.Errorf("could not get storage key for storage account %s: %v", accountName, err)
	}

	return accountName, accountKey, nil
}

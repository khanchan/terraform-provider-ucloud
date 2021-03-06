package ucloud

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudInstances() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudInstancesRead,

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},

			"tag": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"instances": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zone": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"cpu": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"memory": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"instance_type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"charge_type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"auto_renew": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"remark": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"tag": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"create_time": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"expire_time": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"disk_set": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"size": {
										Type:     schema.TypeInt,
										Computed: true,
									},

									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"is_boot": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},

						"ip_set": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ip": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"internet_type": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceUCloudInstancesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).uhostconn

	req := conn.NewDescribeUHostInstanceRequest()

	if ids, ok := d.GetOk("ids"); ok {
		req.UHostIds = schemaSetToStringSlice(ids)
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		req.Zone = ucloud.String(v.(string))
	}

	if v, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(v.(string))
	}

	var instances []uhost.UHostInstanceSet
	var limit int = 100
	var offset int
	var totalCount int
	for {
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := conn.DescribeUHostInstance(req)
		if err != nil {
			return fmt.Errorf("error on reading instance list, %s", err)
		}

		if resp == nil || len(resp.UHostSet) < 1 {
			break
		}

		instances = append(instances, resp.UHostSet...)

		totalCount = totalCount + resp.TotalCount

		if len(resp.UHostSet) < limit {
			break
		}

		offset = offset + limit
	}

	d.Set("total_count", totalCount)
	err := dataSourceUCloudInstancesSave(d, instances)
	if err != nil {
		return fmt.Errorf("error on reading instance list, %s", err)
	}

	return nil
}

func dataSourceUCloudInstancesSave(d *schema.ResourceData, instances []uhost.UHostInstanceSet) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, instance := range instances {
		ids = append(ids, string(instance.UHostId))

		ipSet := []map[string]interface{}{}
		for _, item := range instance.IPSet {
			ipSet = append(ipSet, map[string]interface{}{
				"ip":            item.IP,
				"internet_type": item.Type,
			})
		}

		diskSet := []map[string]interface{}{}
		for _, item := range instance.DiskSet {
			diskSet = append(diskSet, map[string]interface{}{
				"type":    item.DiskType,
				"size":    item.Size,
				"id":      item.DiskId,
				"is_boot": item.IsBoot,
			})
		}

		data = append(data, map[string]interface{}{
			"availability_zone": instance.Zone,
			"id":                instance.UHostId,
			"name":              instance.Name,
			"cpu":               instance.CPU,
			"memory":            instance.Memory,
			"instance_type":     instanceTypeSetFunc(instance.CPU, instance.Memory/1024),
			"create_time":       timestampToString(instance.CreateTime),
			"expire_time":       timestampToString(instance.ExpireTime),
			"auto_renew":        instance.AutoRenew,
			"remark":            instance.Remark,
			"tag":               instance.Tag,
			"status":            instance.State,
			"charge_type":       instance.ChargeType,
			"ip_set":            ipSet,
			"disk_set":          diskSet,
		})
	}

	d.SetId(hashStringArray(ids))
	if err := d.Set("instances", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}

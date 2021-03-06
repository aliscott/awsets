package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/resource"
)

type AWSCloudtrailTrail struct {
}

func init() {
	i := AWSCloudtrailTrail{}
	listers = append(listers, i)
}

func (l AWSCloudtrailTrail) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.CloudtrailTrail}
}

func (l AWSCloudtrailTrail) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := cloudtrail.New(ctx.AWSCfg)
	rg := resource.NewGroup()

	trails, err := svc.DescribeTrailsRequest(&cloudtrail.DescribeTrailsInput{
		IncludeShadowTrails: aws.Bool(true),
	}).Send(ctx.Context)
	if err != nil {
		return rg, fmt.Errorf("failed to list iot thing: %w", err)
	}
	for _, trail := range trails.TrailList {
		r := resource.New(ctx, resource.CloudtrailTrail, trail.Name, trail.Name, trail)
		r.AddRelation(resource.KmsKey, trail.KmsKeyId, "")
		if trail.S3BucketName != nil {
			r.AddRelation(resource.S3Bucket, trail.S3BucketName, "")
		}
		if trail.SnsTopicARN != nil {
			snsArn := arn.ParseP(trail.SnsTopicARN)
			r.AddRelation(resource.SnsTopic, snsArn.ResourceId, snsArn.ResourceVersion)
		}
		//if trail.CloudWatchLogsLogGroupArn != nil { //TODO re-enable after arn parsing is fixed for log groups
		//	cwLogsArn := arn.ParseP(trail.CloudWatchLogsLogGroupArn)
		//	r.AddRelation(resource.LogGroup, cwLogsArn.ResourceId, "")
		//}
		if trail.CloudWatchLogsRoleArn != nil {
			cwLogsRoleArn := arn.ParseP(trail.CloudWatchLogsRoleArn)
			r.AddRelation(resource.LogGroup, cwLogsRoleArn.ResourceId, "")
		}
		if trail.HomeRegion != nil && *trail.HomeRegion == ctx.Region() {
			statusRes, err := svc.GetTrailStatusRequest(&cloudtrail.GetTrailStatusInput{
				Name: trail.Name,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to get status for trail %s: %w", *trail.Name, err)
			}
			r.AddAttribute("Status", statusRes.GetTrailStatusOutput)
		}
		rg.AddResource(r)
	}
	return rg, nil
}
